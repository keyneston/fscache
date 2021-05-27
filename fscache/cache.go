package fscache

import (
	"context"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/keyneston/fscachemonitor/fslist"
	"github.com/keyneston/fscachemonitor/internal/shared"
	"github.com/keyneston/fscachemonitor/proto"
	"github.com/keyneston/fscachemonitor/watcher"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ proto.FSCacheServer = &FSCache{}

type FSCache struct {
	proto.UnimplementedFSCacheServer

	Root string

	fileList fslist.FSList
	watcher  watcher.Watcher
	socket   net.Listener
	server   *grpc.Server

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce *sync.Once

	logger *logrus.Logger
}

func New(socketLocation, root string, mode fslist.Mode) (*FSCache, error) {
	watcher, err := watcher.New(root)
	if err != nil {
		return nil, err
	}

	socket, err := net.Listen("unix", socketLocation)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	fs := &FSCache{
		Root:      root,
		watcher:   watcher,
		socket:    socket,
		logger:    shared.Logger().WithField("object", "fscache").Logger,
		server:    grpc.NewServer(),
		cancel:    cancel,
		ctx:       ctx,
		closeOnce: &sync.Once{},
	}

	proto.RegisterFSCacheServer(fs.server, fs)

	fs.fileList, err = fslist.New(mode)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *FSCache) Run() {
	go fs.server.Serve(fs.socket)

	fs.setSignalHandlers()
	fs.watcher.Start()
	fs.init()

	for {
		select {
		case events := <-fs.watcher.Stream():
			for _, e := range events {
				fs.handleEvent(e)
			}
		case <-fs.ctx.Done():
			fs.logger.WithError(fs.ctx.Err()).Warn("Receive context.Done")
			return
		}
	}
}

func (fs *FSCache) setSignalHandlers() {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		fs.Close()
	}()
}

func eventToAddData(e watcher.Event) fslist.AddData {
	return fslist.AddData{
		Name:      e.Path,
		IsDir:     e.Dir,
		UpdatedAt: time.Now(),
	}
}

func (fs *FSCache) handleEvent(e watcher.Event) {
	// TODO: find a better way:
	for _, seg := range strings.Split(e.Path, "/") {
		if skipFile(seg) {
			fs.logger.Debugf("Skipping %q", e.Path)
			return
		}
	}

	switch e.Type {
	case watcher.EventTypeDelete:
		fs.logger.Tracef("Removing %q", e.Path)
		if err := fs.fileList.Delete(eventToAddData(e)); err != nil {
			fs.logger.Errorf("Error deleting file: %v", err)
		}
	case watcher.EventTypeAdd:
		fs.logger.Tracef("Adding %q", e.Path)
		if err := fs.fileList.Add(eventToAddData(e)); err != nil {
			fs.logger.Errorf("Error adding file: %v", err)
		}
	}
}

func (fs *FSCache) Close() {
	fs.closeOnce.Do(func() {
		fs.logger.Warn("Received stop, shutting down")
		fs.watcher.Stop()
		fs.cancel()
		go fs.server.GracefulStop()
	})
}

// init does the initial setup of walking
func (fs *FSCache) init() {
	filepath.WalkDir(fs.Root, func(path string, d os.DirEntry, err error) error {
		select {
		case <-fs.ctx.Done():
			return fs.ctx.Err()
		default:
		}

		if skipFile(path) {
			shared.Logger().Debugf("Skipping %q", path)
			return filepath.SkipDir
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		fs.fileList.Add(fslist.AddData{
			Name:      path,
			UpdatedAt: info.ModTime(),
			IsDir:     d.IsDir(),
		})
		return nil
	})
}

func skipFile(path string) bool {
	switch filepath.Base(path) {
	case ".git", ".svn":
		return true
	default:
		return false
	}
}

func (fs *FSCache) GetFiles(req *proto.ListRequest, srv proto.FSCache_GetFilesServer) error {
	fs.logger.WithField("req", req).Debugf("Received request")

	opts := fslist.ReadOptions{
		DirsOnly: req.DirsOnly,
		Prefix:   req.Prefix,
		Limit:    int(req.Limit),
	}

	for file := range fs.fileList.Fetch(opts) {
		f := &proto.File{
			Name: file.Name,
		}

		if err := srv.Send(f); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FSCache) Shutdown(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	fs.Close()
	return req, nil
}

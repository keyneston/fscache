package fscache

import (
	"bytes"
	"context"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/keyneston/fscache/fslist"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/keyneston/fscache/proto"
	"github.com/keyneston/fscache/watcher"
	"github.com/monochromegane/go-gitignore"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var DefaultFlushTime = time.Second * 1

var _ proto.FSCacheServer = &FSCache{}

const watchIgnores = `
.git/
.svn/
Application Support/
.cache/
.DS_File
pkg/mod
pkg/sumdb
pkg/mod
`

type FSCache struct {
	proto.UnimplementedFSCacheServer

	Root string

	fileList fslist.FSList
	watcher  watcher.Watcher
	socket   net.Listener
	server   *grpc.Server
	ignore   gitignore.IgnoreMatcher

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
		ignore:    gitignore.NewGitIgnoreFromReader("/", bytes.NewBufferString(watchIgnores)),
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

	flushTick := time.NewTicker(DefaultFlushTime)

	for {
		select {
		case events := <-fs.watcher.Stream():
			for _, e := range events {
				fs.handleEvent(e)
			}
		case <-fs.ctx.Done():
			fs.logger.WithError(fs.ctx.Err()).Warn("Receive context.Done")
			return
		case <-flushTick.C:
			if err := fs.fileList.Flush(); err != nil {
				fs.logger.WithError(err).Error("error flushing fslist")
			}
		}
	}
}

func (fs *FSCache) Flush() error {
	return fs.fileList.Flush()
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

func checkSkipPath(ignore gitignore.IgnoreMatcher, path string, dir bool) bool {
	// TODO: find a better way:
	segments := strings.Split(path, "/")
	for i := range segments {
		if i != 0 {
			dir = true
		}

		path = strings.Join(segments[0:i], "/")
		if ignore.Match(path, dir) {
			return true
		}
	}

	return false
}

func (fs *FSCache) handleEvent(e watcher.Event) {
	if checkSkipPath(fs.ignore, e.Path, e.Dir) {
		fs.logger.Debugf("Skipping %#q", e.Path)
		return
	}

	switch e.Type {
	case watcher.EventTypeDelete:
		fs.logger.Tracef("Removing %#q", e.Path)
		if err := fs.fileList.Delete(eventToAddData(e)); err != nil {
			fs.logger.Errorf("Error deleting file: %v", err)
		}
	case watcher.EventTypeAdd:
		fs.logger.Tracef("Adding %#q", e.Path)
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
		fs.fileList.Close()
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

		if fs.ignore.Match(path, d.IsDir()) {
			shared.Logger().Debugf("Skipping %q", path)
			return filepath.SkipDir
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		return fs.fileList.Add(fslist.AddData{
			Name:      abs,
			UpdatedAt: info.ModTime(),
			IsDir:     d.IsDir(),
		})
	})
}

func (fs *FSCache) GetFiles(req *proto.ListRequest, srv proto.FSCache_GetFilesServer) error {
	fs.logger.WithField("req", req).Debugf("Received request")

	opts := fslist.ReadOptions{
		DirsOnly:   req.DirsOnly,
		Prefix:     req.Prefix,
		Limit:      int(req.Limit),
		CurrentDir: req.CurrentDir,
	}

	batchSize := 10
	if req.BatchSize != 0 {
		batchSize = int(req.BatchSize)
	}

	files := &proto.Files{}
	for file := range fs.fileList.Fetch(opts) {
		f := &proto.File{
			Name: file.Name,
			Dir:  file.IsDir,
		}

		files.Files = append(files.Files, f)

		if len(files.Files) >= batchSize {
			if err := srv.Send(files); err != nil {
				return err
			}
			files = &proto.Files{}
		}
	}

	// Send any remaining data:
	if len(files.Files) > 0 {
		if err := srv.Send(files); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FSCache) Shutdown(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	fs.Close()
	return req, nil
}

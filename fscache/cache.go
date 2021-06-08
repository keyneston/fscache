package fscache

import (
	"context"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/keyneston/fscache/fslist"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/keyneston/fscache/proto"
	"github.com/keyneston/fscache/watcher"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var DefaultFlushTime = time.Second * 1

var _ proto.FSCacheServer = &FSCache{}

type FSCache struct {
	proto.UnimplementedFSCacheServer

	Root string

	fileList fslist.FSList
	watcher  watcher.Watcher
	socket   net.Listener
	server   *grpc.Server
	ignore   GlobalIgnore

	ctx           context.Context
	cancel        context.CancelFunc
	closeOnce     *sync.Once
	signalRestart bool
	wg            *sync.WaitGroup

	logger zerolog.Logger
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
		logger:    shared.Logger().With().Str("object", "fscache").Logger(),
		server:    grpc.NewServer(),
		cancel:    cancel,
		ctx:       ctx,
		closeOnce: &sync.Once{},
		wg:        &sync.WaitGroup{},
		ignore:    NewGlobalIgnore(root),
	}

	proto.RegisterFSCacheServer(fs.server, fs)

	fs.fileList, err = fslist.New(mode)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

// Run runs the main loop. It returns true if the server should restart instead
// of shutting down.
func (fs *FSCache) Run() bool {
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
			fs.logger.Warn().Err(fs.ctx.Err()).Msg("receive context.Done")
			return fs.signalRestart
		case <-flushTick.C:
			if err := fs.fileList.Flush(); err != nil {
				fs.logger.Error().Err(err).Msg("error flushing fslist")
			}
		}
	}

	return fs.signalRestart
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
	updatedAt := time.Now()

	return fslist.AddData{
		Name:      e.Path,
		IsDir:     e.Dir,
		UpdatedAt: &updatedAt,
	}
}

func (fs *FSCache) handleEvent(e watcher.Event) {
	if fs.ignore.Match(e.Path, e.Dir) {
		fs.logger.Debug().Msgf("Skipping %#q", e.Path)
		return
	}

	switch e.Type {
	case watcher.EventTypeDelete:
		fs.logger.Trace().Str("path", e.Path).Msg("removing")
		if err := fs.fileList.Delete(eventToAddData(e)); err != nil {
			fs.logger.Error().Str("path", e.Path).Err(err).Msgf("Error deleting file: %v", err)
		}
	case watcher.EventTypeAdd:
		fs.logger.Trace().Str("path", e.Path).Msg("adding")
		if err := fs.fileList.Add(eventToAddData(e)); err != nil {
			fs.logger.Error().Str("path", e.Path).Err(err).Msgf("Error adding file: %v", err)
		}
	}
}

func (fs *FSCache) Close() {
	fs.closeOnce.Do(func() {
		fs.logger.Warn().Msg("Received stop, shutting down")
		fs.watcher.Stop()
		fs.cancel()
		go fs.server.GracefulStop()

		// Wait for wg to finish before closing the database.
		fs.wg.Wait()

		fs.fileList.Close()
	})
}

// init does the initial setup of walking
func (fs *FSCache) init() {
	fs.wg.Add(1)
	defer fs.wg.Done()

	if entry, err := getDirEntry(fs.Root); err != nil {
		fs.logger.Error().Err(err).Str("root", fs.Root).Msg("error getting entry for root")
	} else {
		// first "walk" the root directory itself
		if err := fs.walkFunc(fs.Root, entry, nil); err != nil {
			fs.logger.Error().Err(err).Msg("error walking root")
		}
	}

	filepath.WalkDir(fs.Root, fs.walkFunc)
}

func (fs *FSCache) walkFunc(path string, d os.DirEntry, err error) error {
	select {
	case <-fs.ctx.Done():
		return fs.ctx.Err()
	default:
	}

	isDir := false
	updatedAt := time.Time{}
	if d != nil {
		isDir = d.IsDir()

		if info, err := d.Info(); err == nil {
			updatedAt = info.ModTime().UTC()
		}
	}

	if fs.ignore.Match(path, isDir) {
		fs.logger.Debug().Str("path", path).Msgf("Skipping %q", path)
		if d.IsDir() {
			return filepath.SkipDir
		} else {
			return nil
		}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	return fs.fileList.Add(fslist.AddData{
		Name:      abs,
		UpdatedAt: &updatedAt,
		IsDir:     isDir,
	})
}

func (fs *FSCache) GetFiles(req *proto.ListRequest, srv proto.FSCache_GetFilesServer) error {
	fs.logger.Debug().Interface("req", req).Msg("Received request")

	opts := fslist.ReadOptions{
		DirsOnly:   req.DirsOnly,
		FilesOnly:  req.FilesOnly,
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

func (fs *FSCache) Shutdown(ctx context.Context, req *proto.ShutdownRequest) (*emptypb.Empty, error) {
	fs.wg.Add(1)
	defer fs.wg.Done()

	if req != nil && req.Restart {
		fs.signalRestart = true
	}

	go fs.Close()
	return &emptypb.Empty{}, nil
}

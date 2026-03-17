package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirbyevanj/kvqtool-api-server/internal/config"
	"github.com/kirbyevanj/kvqtool-api-server/internal/service"
	"github.com/kirbyevanj/kvqtool-api-server/internal/storage"
	transport "github.com/kirbyevanj/kvqtool-api-server/internal/transport/http"
	"github.com/valyala/fasthttp"
)

func main() {
	cfg := config.Load()
	logger := initLogger(cfg)

	db, err := storage.NewPostgresDB(cfg.DSN, logger)
	if err != nil {
		logger.Error("postgres init failed", "err", err)
		os.Exit(1)
	}

	if err := storage.RunMigrations(context.Background(), db, logger); err != nil {
		logger.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	s3Client, err := storage.NewS3Client(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Region, cfg.S3PublicEndpoint, logger)
	if err != nil {
		logger.Error("s3 init failed", "err", err)
		os.Exit(1)
	}

	temporalClient, err := storage.NewTemporalClient(cfg.TemporalHost, logger)
	if err != nil {
		logger.Error("temporal init failed", "err", err)
		os.Exit(1)
	}
	defer temporalClient.Close()

	projects := service.NewProjectService(db, s3Client, logger)
	folders := service.NewFolderService(db, logger)
	resources := service.NewResourceService(db, s3Client, logger)
	workflows := service.NewWorkflowService(db, logger)
	jobs := service.NewJobService(db, temporalClient, logger)

	srv := transport.NewServer(logger, projects, folders, resources, workflows, jobs, temporalClient)

	logger.Info("starting api-server", "addr", cfg.ListenAddr)

	go func() {
		if err := fasthttp.ListenAndServe(cfg.ListenAddr, srv.Handler()); err != nil {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logger.Info("shutting down")
}

func initLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.IsDev() {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

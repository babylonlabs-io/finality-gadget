package server

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/babylonlabs-io/finality-gadget/config"
	"github.com/babylonlabs-io/finality-gadget/db"
	"github.com/babylonlabs-io/finality-gadget/finalitygadget"
	"github.com/babylonlabs-io/finality-gadget/proto"
	"github.com/lightningnetwork/lnd/signal"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

// Server is the main daemon construct for the finality gadget server. It
// handles spinning up both the gRPC and HTTP servers, the database, and any
// other components that the the finality gadget server needs to run.
type Server struct {
	proto.UnimplementedFinalityGadgetServer
	fg  finalitygadget.IFinalityGadget
	cfg *config.Config
	db  db.IDatabaseHandler

	logger *zap.Logger

	interceptor signal.Interceptor
	started     int32
}

// NewFinalityGadgetServer creates a new server with the given config.
func NewFinalityGadgetServer(cfg *config.Config, db db.IDatabaseHandler, fg finalitygadget.IFinalityGadget, sig signal.Interceptor, logger *zap.Logger) *Server {
	return &Server{
		fg:          fg,
		cfg:         cfg,
		db:          db,
		logger:      logger,
		interceptor: sig,
	}
}

// RunUntilShutdown runs the main finality gadget server loop until a signal is
// received to shut down the process.
func (s *Server) RunUntilShutdown() error {
	if atomic.AddInt32(&s.started, 1) != 1 {
		return nil
	}

	defer func() {
		s.logger.Info("Closing database...")
		if err := s.db.Close(); err != nil {
			s.logger.Error("Failed to close database", zap.Error(err))
		} else {
			s.logger.Info("Database closed")
		}
	}()

	if err := s.startGrpcServer(); err != nil {
		return fmt.Errorf("failed to start gRPC listener: %v", err)
	}

	if err := s.startHttpServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %v", err)
	}

	s.logger.Info("Finality gadget is active")

	// Wait for shutdown signal from either a graceful server stop or from
	// the interrupt handler.
	<-s.interceptor.ShutdownChannel()

	return nil
}

func (s *Server) startGrpcServer() error {
	lis, err := net.Listen("tcp", s.cfg.GRPCListener)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.cfg.GRPCListener, err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	if err := s.RegisterWithGrpcServer(grpcServer); err != nil {
		return fmt.Errorf("failed to register gRPC server: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.logger.Info("RPC server listening", zap.String("address", lis.Addr().String()))

		// Close the ready chan to indicate we are listening.
		defer lis.Close()

		wg.Done()
		_ = grpcServer.Serve(lis)
	}()
	wg.Wait()
	return nil
}

func (s *Server) startHttpServer() error {
	corsOpts := cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
	}

	httpServer := &http.Server{
		Addr:              s.cfg.HTTPListener,
		Handler:           cors.New(corsOpts).Handler(s.newHttpHandler()),
		ReadHeaderTimeout: 30 * time.Second,
	}

	s.logger.Info("Starting standalone HTTP server", zap.String("address", s.cfg.HTTPListener))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed: %w", err)
	}
	return nil
}

package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/babylonchain/babylon-finality-gadget/db"
	"github.com/babylonchain/babylon-finality-gadget/finalitygadget"
	"github.com/babylonchain/babylon-finality-gadget/finalitygadget/config"
	"github.com/lightningnetwork/lnd/signal"

	"google.golang.org/grpc"
)

// Server is the main daemon construct for the EOTS manager server. It handles
// spinning up the RPC sever, the database, and any other components that the
// EOTS manager server needs to function.
type Server struct {
	started int32

	rpcServer   *rpcServer
	cfg         *config.Config
	db          *db.BBoltHandler
	interceptor signal.Interceptor

	quit chan struct{}
}

// NewEOTSManagerServer creates a new server with the given config.
func NewFinalityGadgetServer(cfg *config.Config, db *db.BBoltHandler, fg finalitygadget.IFinalityGadget, sig signal.Interceptor) *Server {
	return &Server{
		cfg:         cfg,
		rpcServer:   newRPCServer(fg),
		db:          db,
		interceptor: sig,
		quit:        make(chan struct{}, 1),
	}
}

// RunUntilShutdown runs the main finality gadget server loop until a signal is
// received to shut down the process.
func (s *Server) RunUntilShutdown() error {
	if atomic.AddInt32(&s.started, 1) != 1 {
		return nil
	}

	defer func() {
		log.Printf("Closing database...")
		s.db.Close()
		log.Printf("Database closed")
	}()

	listenAddr := "localhost:" + s.cfg.GRPCServerPort
	// we create listeners from the RPCListeners defined
	// in the config.
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	if err := s.rpcServer.RegisterWithGrpcServer(grpcServer); err != nil {
		return fmt.Errorf("failed to register gRPC server: %w", err)
	}

	// All the necessary components have been registered, so we can
	// actually start listening for requests.
	if err := s.startGrpcListen(grpcServer, []net.Listener{lis}); err != nil {
		return fmt.Errorf("failed to start gRPC listener: %v", err)
	}

	log.Printf("Finality gadget is active")

	// Wait for shutdown signal from either a graceful server stop or from
	// the interrupt handler.
	<-s.interceptor.ShutdownChannel()

	return nil
}

// startGrpcListen starts the GRPC server on the passed listeners.
func (s *Server) startGrpcListen(grpcServer *grpc.Server, listeners []net.Listener) error {

	// Use a WaitGroup so we can be sure the instructions on how to input the
	// password is the last thing to be printed to the console.
	var wg sync.WaitGroup

	for _, lis := range listeners {
		wg.Add(1)
		go func(lis net.Listener) {
			log.Printf("RPC server listening: %s", lis.Addr().String())

			// Close the ready chan to indicate we are listening.
			defer lis.Close()

			wg.Done()
			_ = grpcServer.Serve(lis)
		}(lis)
	}

	// Wait for gRPC servers to be up running.
	wg.Wait()

	return nil
}

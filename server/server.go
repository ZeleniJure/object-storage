package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Server struct {
	Routes *mux.Router
	Ctx    context.Context
}

func New() Server {
	s := Server{}

	log.Info().Int("pid", os.Getpid()).Int("uid", os.Getuid()).Int("gid", os.Getgid()).Msg("Server started")

	timeout := time.Duration(viper.GetInt("server.timeout")) * time.Second
	s.Routes = NewRouter()
	srv := &http.Server{
		Addr:    viper.GetString("server.address"),
		Handler: s.Routes,
		// ReadHeaderTimeout is here as well
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout,
	}

	shutdownCtx := s.gracefullShutdown(srv)
	s.Ctx = shutdownCtx

	go func() {
		log.Info().Str("address", srv.Addr).Msg("Server listening")
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("Server stopped")
			s.triggerShutdown(shutdownCtx, srv)
		}
	}()

	return s
}

func (s *Server) gracefullShutdown(server *http.Server) context.Context {
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		// Shutdown signal with grace period of X seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, viper.GetDuration("server.shutdown")*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal().Msgf("graceful shutdown timed out.. forcing exit.")
			}
		}()

		s.triggerShutdown(shutdownCtx, server)
		serverStopCtx()
	}()
	return serverCtx
}

func (s *Server) triggerShutdown(ctx context.Context, server *http.Server) {
	err := server.Shutdown(ctx)
	if err != nil {
		log.Error().Stack().Msgf("error shutting down server (%s): %v", server.Addr, err)
		err = server.Close()
		if err != nil {
			log.Error().Stack().Msgf("error closing server (%s): %v", server.Addr, err)
		}
	}
}

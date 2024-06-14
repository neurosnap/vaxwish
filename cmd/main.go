package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/neurosnap/vaxwish"
)

func main() {
	GitSshServer()
}

func authHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}

func GitSshServer() {
	host := os.Getenv("SSH_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("SSH_PORT")
	if port == "" {
		port = "2222"
	}

	logger := slog.Default()

	vx, err := vaxis.New(vaxis.Options{
		DisableMouse: true,
	})
	if err != nil {
		panic(err)
	}
	defer vx.Close()

	s, err := wish.NewServer(
		wish.WithAddress(
			fmt.Sprintf("%s:%s", host, port),
		),
		wish.WithHostKeyPath(
			filepath.Join("ssh_data", "term_info_ed25519"),
		),
		wish.WithPublicKeyAuth(authHandler),
		wish.WithMiddleware(
			vaxwish.VaxisMiddleware(vx),
		),
	)

	if err != nil {
		logger.Error("could not create server", "err", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			logger.Error("serve error", "err", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		logger.Error("shutdown", "err", err)
		os.Exit(1)
	}
}

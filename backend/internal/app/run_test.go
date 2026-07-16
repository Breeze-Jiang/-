package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestRunStopsBothServersWhenContextIsCancelled(t *testing.T) {
	application := &App{
		server:      &http.Server{Addr: "127.0.0.1:0", Handler: http.NewServeMux()},
		adminServer: &http.Server{Addr: "127.0.0.1:0", Handler: http.NewServeMux()},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := application.Run(ctx, time.Second); err != nil {
		t.Fatalf("run after context cancellation: %v", err)
	}
}

func TestRunShutsDownPeerWhenAListenerFails(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	application := &App{
		server:      &http.Server{Addr: listener.Addr().String(), Handler: http.NewServeMux()},
		adminServer: &http.Server{Addr: "127.0.0.1:0", Handler: http.NewServeMux()},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = application.Run(ctx, time.Second)
	if err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected listener failure, got %v", err)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"agentrun/internal/egress"
)

const (
	readHeaderTimeout = 10 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 5 * time.Second
)

type allowList []string

func (a *allowList) String() string {
	return strings.Join(*a, ",")
}

func (a *allowList) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stderr))
}

func run(ctx context.Context, args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("agentrun-egress", flag.ContinueOnError)
	fs.SetOutput(stderr)
	listen := fs.String("listen", ":3128", "address to listen on")
	var allow allowList
	fs.Var(&allow, "allow", "host:port the proxy may tunnel to; repeatable")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(allow) == 0 {
		fmt.Fprintln(stderr, "agentrun-egress: at least one -allow host:port is required")
		return 2
	}

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := serve(ctx, ln, newHandler(allow, stderr)); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func newHandler(allow []string, stderr io.Writer) http.Handler {
	return egress.New(allow, egress.WithLog(func(decision, target string) {
		fmt.Fprintf(stderr, "egress %s %s\n", decision, target)
	}))
}

func serve(ctx context.Context, ln net.Listener, handler http.Handler) error {
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

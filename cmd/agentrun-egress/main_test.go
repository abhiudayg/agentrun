package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const ioDeadline = 5 * time.Second

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func startEcho(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				io.Copy(conn, conn)
			}()
		}
	}()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	return port
}

func connect(t *testing.T, proxyAddr, target string) *http.Response {
	t.Helper()
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := conn.SetDeadline(time.Now().Add(ioDeadline)); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", target, target); err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestRunRefusesToStartWithoutAllow(t *testing.T) {
	var stderr bytes.Buffer

	code := run(context.Background(), []string{"-listen", "127.0.0.1:0"}, &stderr)

	if code == 0 {
		t.Fatalf("run() = 0, want non-zero exit")
	}
	if !strings.Contains(stderr.String(), "-allow") {
		t.Fatalf("stderr = %q, want a message naming -allow", stderr.String())
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stderr bytes.Buffer

	code := run(context.Background(), []string{"-bogus"}, &stderr)

	if code == 0 {
		t.Fatalf("run() = 0, want non-zero exit")
	}
}

func TestRunExitsZeroAfterShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stderr bytes.Buffer

	code := run(ctx, []string{"-listen", "127.0.0.1:0", "-allow", "example.com:443"}, &stderr)

	if code != 0 {
		t.Fatalf("run() = %d, want 0 (stderr %q)", code, stderr.String())
	}
}

func TestDeniedConnectLogsLineAndReturnsForbidden(t *testing.T) {
	var stderr syncBuffer
	srv := httptest.NewServer(newHandler([]string{"localhost:443"}, &stderr))
	t.Cleanup(srv.Close)

	resp := connect(t, srv.Listener.Addr().String(), "example.com:443")

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	if got, want := stderr.String(), "egress denied example.com:443\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestAllowedConnectLogsLineAndTunnels(t *testing.T) {
	echoPort := startEcho(t)
	target := "localhost:" + echoPort
	var stderr syncBuffer
	srv := httptest.NewServer(newHandler([]string{target}, &stderr))
	t.Cleanup(srv.Close)
	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := conn.SetDeadline(time.Now().Add(ioDeadline)); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", target, target); err != nil {
		t.Fatal(err)
	}
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	msg := "ping"
	if _, err := io.WriteString(conn, msg); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, len(msg))
	if _, err := io.ReadFull(br, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != msg {
		t.Fatalf("echoed %q, want %q", got, msg)
	}
	if want := "egress allowed " + target + "\n"; stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
}

func TestServeReturnsNilWhenContextIsCancelled(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- serve(ctx, ln, http.NotFoundHandler())
	}()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve() = %v, want nil", err)
		}
	case <-time.After(ioDeadline):
		t.Fatal("serve() did not return after context cancellation")
	}
}

package egress

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const ioDeadline = 5 * time.Second

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

func closedPort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()
	return port
}

func startProxy(t *testing.T, allowed ...string) string {
	t.Helper()
	srv := httptest.NewServer(New(allowed))
	t.Cleanup(srv.Close)
	return srv.Listener.Addr().String()
}

func connect(t *testing.T, proxyAddr, target string) (net.Conn, *bufio.Reader, *http.Response) {
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
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	return conn, br, resp
}

func TestConnectToAllowedHostTunnelsBytesBothWays(t *testing.T) {
	echoPort := startEcho(t)
	proxyAddr := startProxy(t, "localhost:"+echoPort)

	conn, br, resp := connect(t, proxyAddr, "localhost:"+echoPort)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	msg := "hello through the tunnel"
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
}

func TestConnectToHostNotOnAllowlistIsForbidden(t *testing.T) {
	echoPort := startEcho(t)
	otherPort := startEcho(t)
	proxyAddr := startProxy(t, "localhost:"+echoPort)

	tests := []struct {
		name   string
		target string
	}{
		{name: "unlisted host name", target: "example.com:443"},
		{name: "subdomain of allowed host", target: "sub.localhost:" + echoPort},
		{name: "IPv4 literal of a live listener", target: "127.0.0.1:" + echoPort},
		{name: "IPv6 literal", target: "[::1]:" + echoPort},
		{name: "allowed host on an unlisted port", target: "localhost:" + otherPort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, resp := connect(t, proxyAddr, tt.target)
			if resp.StatusCode != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
			}
		})
	}
}

func TestNonConnectMethodIsNotAllowed(t *testing.T) {
	proxyAddr := startProxy(t, "localhost:443")

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, "http://"+proxyAddr+"/", nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestMalformedConnectTargetIsBadRequest(t *testing.T) {
	proxyAddr := startProxy(t, "localhost:443")

	targets := []string{
		"localhost",
		"localhost:",
		":443",
		"localhost:https",
		"localhost:0",
		"localhost:70000",
	}
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			_, _, resp := connect(t, proxyAddr, target)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
			}
		})
	}
}

func TestAllowedHostWithClosedPortIsBadGateway(t *testing.T) {
	port := closedPort(t)
	proxyAddr := startProxy(t, "localhost:"+port)

	_, _, resp := connect(t, proxyAddr, "localhost:"+port)
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}
}

func TestHostMatchingIgnoresCaseAndPort(t *testing.T) {
	echoPort := startEcho(t)
	proxyAddr := startProxy(t, "LocalHost:"+echoPort)

	_, _, resp := connect(t, proxyAddr, "LOCALHOST:"+echoPort)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

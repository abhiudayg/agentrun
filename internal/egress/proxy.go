package egress

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dialTimeout = 10 * time.Second

type proxy struct {
	hosts map[string]struct{}
}

func New(allowed []string) http.Handler {
	hosts := make(map[string]struct{}, len(allowed))
	for _, host := range allowed {
		hosts[strings.ToLower(host)] = struct{}{}
	}
	return &proxy{hosts: hosts}
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	host, port, err := net.SplitHostPort(r.Host)
	if err != nil || host == "" || !validPort(port) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if !p.allowed(host) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	upstream, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), dialTimeout)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}
	defer upstream.Close()

	client, brw, err := http.NewResponseController(w).Hijack()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	if _, err := io.WriteString(brw, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		return
	}
	if err := brw.Flush(); err != nil {
		return
	}
	tunnel(upstream, client, brw.Reader)
}

func (p *proxy) allowed(host string) bool {
	if net.ParseIP(host) != nil {
		return false
	}
	_, ok := p.hosts[strings.ToLower(host)]
	return ok
}

func validPort(port string) bool {
	n, err := strconv.ParseUint(port, 10, 16)
	return err == nil && n > 0
}

func tunnel(upstream, client net.Conn, clientReader *bufio.Reader) {
	done := make(chan struct{}, 2)
	go func() {
		io.Copy(upstream, clientReader)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(client, upstream)
		done <- struct{}{}
	}()
	<-done
}

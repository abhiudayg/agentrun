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
	log   func(decision, target string)
}

type Option func(*proxy)

func WithLog(fn func(decision, target string)) Option {
	return func(p *proxy) { p.log = fn }
}

func New(allowed []string, opts ...Option) http.Handler {
	hosts := make(map[string]struct{}, len(allowed))
	for _, authority := range allowed {
		hosts[strings.ToLower(authority)] = struct{}{}
	}
	p := &proxy{hosts: hosts, log: func(string, string) {}}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	host, port, err := net.SplitHostPort(r.Host)
	if err != nil || host == "" || !validPort(port) {
		p.log("denied", r.Host)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	target := net.JoinHostPort(host, port)
	if _, ok := p.hosts[strings.ToLower(target)]; !ok {
		p.log("denied", target)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	p.log("allowed", target)

	upstream, err := net.DialTimeout("tcp", target, dialTimeout)
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

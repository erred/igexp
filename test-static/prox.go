package prox

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"sync"
)

var s *sync.Once
var cmd *exec.Cmd
var rp *httputil.ReverseProxy

// P proxies requests to compiled_binary
func P(w http.ResponseWriter, r *http.Request) {
	// setup
	s.Do(func() {
		cmd = exec.Command("compiled_binary")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
		url, err := url.Parse("localhost:8080")
		if err != nil {
			// TODO handle error
		}
		rp = httputil.NewSingleHostReverseProxy(url)
	})

	rp.ServeHTTP(w, r)
}

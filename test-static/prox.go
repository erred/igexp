package prox

import (
	"log"
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
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered: ", r)
		}
	}()
	// setup
	s.Do(func() {
		log.Println("initializing...")
		cmd = exec.Command("/srv/files/compiled_binary")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			log.Println(err)
		}
		url, err := url.Parse("localhost:8080")
		if err != nil {
			log.Println(err)
		}
		rp = httputil.NewSingleHostReverseProxy(url)
	})

	rp.ServeHTTP(w, r)
}

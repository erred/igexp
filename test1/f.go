package test1

import (
	"net/http"
	"os"
	"strings"
)

func F(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strings.Join(os.Environ(), "\n")))
}

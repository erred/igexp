package test1

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func F(w http.ResponseWriter, r *http.Request) {
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

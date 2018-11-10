package deploy

import (
	"net/http"
)

func F(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World"))
}

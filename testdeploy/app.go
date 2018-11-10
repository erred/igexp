package deploy

import (
	"net/http"

	"github.com/zmb3/spotify"
)

func F(w http.ResponseWriter, r *http.Request) {
	_ = spotify.Client{}
	w.Write([]byte("Hello, World"))
}

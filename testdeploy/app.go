package deploy

import (
	"net/http"

	"github.com/seankhliao/igtools/goinsta"
)

func F(w http.ResponseWriter, r *http.Request) {
	_ = goinsta.New("", "")
	w.Write([]byte("Hello, World"))
}

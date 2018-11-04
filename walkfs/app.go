package walk

import (
	"net/http"
	"os"
	"path/filepath"
)

func F(w http.ResponseWriter, r *http.Request) {
	filepath.Walk("/", func(path string, info os.FileInfo, err error) error {
		w.Write([]byte(path + "\n"))
		return nil
	})
}

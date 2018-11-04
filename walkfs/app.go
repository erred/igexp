package walk

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func F(w http.ResponseWriter, r *http.Request) {
	paths := []string{}
	filepath.Walk("/", func(path string, info os.FileInfo, err error) error {
		paths = append(paths, path)
		return nil
	})
	w.Write([]byte(strings.Join(paths, "\n")))
}

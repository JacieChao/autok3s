package ui

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/rancher/apiserver/pkg/parse"
)

const defaultPath = "./dist"

func Content() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.FileServer(http.Dir(defaultPath)).ServeHTTP(rw, req)
	})
}

func UI(next http.Handler) http.Handler {
	os.Stat(indexHTML())
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if parse.IsBrowser(req, true) {
			http.ServeFile(resp, req, indexHTML())
		} else {
			next.ServeHTTP(resp, req)
		}
	})
}

func indexHTML() string {
	return filepath.Join(defaultPath, "index.html")
}

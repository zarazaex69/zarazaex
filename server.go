package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed public/*
var content embed.FS

func isCurl(r *http.Request) bool {
	return strings.HasPrefix(r.UserAgent(), "curl/")
}

type curlAwareHandler struct {
	fileServer   http.Handler
	curlResponse []byte
}

func (h *curlAwareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" && isCurl(r) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(h.curlResponse)
		return
	}
	h.fileServer.ServeHTTP(w, r)
}

func main() {
	public, err := fs.Sub(content, "public")
	if err != nil {
		log.Fatal(err)
	}
	curlTxt, err := content.ReadFile("public/curl.txt")
	if err != nil {
		log.Fatal(err)
	}
	handler := &curlAwareHandler{
		fileServer:   http.FileServer(http.FS(public)),
		curlResponse: curlTxt,
	}
	http.Handle("/", handler)
	log.Println("zarazaex running on :8801")
	if err := http.ListenAndServe(":8801", nil); err != nil {
		log.Fatal(err)
	}
}

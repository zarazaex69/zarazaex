package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
)

//go:embed public/*
var content embed.FS

func isCurl(r *http.Request) bool {
	return strings.HasPrefix(r.UserAgent(), "curl/")
}

var compressibleTypes = map[string]bool{
	"text/html":              true,
	"text/css":               true,
	"text/plain":             true,
	"text/xml":               true,
	"text/javascript":        true,
	"application/javascript": true,
	"application/json":       true,
	"application/xml":        true,
	"image/svg+xml":          true,
}

func isCompressible(path string) bool {
	ext := filepath.Ext(path)
	if ext == "" {
		return false
	}
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		return false
	}
	base := ct
	if idx := strings.IndexByte(ct, ';'); idx != -1 {
		base = ct[:idx]
	}
	return compressibleTypes[strings.TrimSpace(base)]
}

func acceptsBrotli(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(part) == "br" || strings.HasPrefix(strings.TrimSpace(part), "br;") {
			return true
		}
	}
	return false
}

type brotliResponseWriter struct {
	http.ResponseWriter
	writer      io.Writer
	wroteHeader bool
}

func (w *brotliResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.writer.Write(b)
}

func (w *brotliResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.ResponseWriter.Header().Del("Content-Length")
	w.ResponseWriter.Header().Set("Content-Encoding", "br")
	w.ResponseWriter.Header().Add("Vary", "Accept-Encoding")
	w.ResponseWriter.WriteHeader(code)
}

func brotliMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !acceptsBrotli(r) || !isCompressible(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		br := brotli.NewWriterLevel(nil, brotli.BestCompression)
		brw := &brotliResponseWriter{
			ResponseWriter: w,
			writer:         br,
		}
		br.Reset(w)

		next.ServeHTTP(brw, r)
		br.Close()
	})
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
	compressed := brotliMiddleware(handler)
	http.Handle("/", compressed)
	log.Println("zarazaex running on :8801")
	if err := http.ListenAndServe(":8801", nil); err != nil {
		log.Fatal(err)
	}
}

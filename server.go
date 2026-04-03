package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

//go:embed public/*
var content embed.FS

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

type cachedFile struct {
	raw         []byte
	compressed  []byte
	contentType string
}

type server struct {
	cache        map[string]*cachedFile
	curlResponse []byte
}

func mimeForPath(path string) string {
	ct := mime.TypeByExtension(filepath.Ext(path))
	if ct == "" {
		return "application/octet-stream"
	}
	return ct
}

func baseType(ct string) string {
	if idx := strings.IndexByte(ct, ';'); idx != -1 {
		return strings.TrimSpace(ct[:idx])
	}
	return ct
}

func compressBrotli(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(len(data) / 2)
	w := brotli.NewWriterLevel(&buf, brotli.BestCompression)
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("brotli write: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("brotli close: %w", err)
	}
	return buf.Bytes(), nil
}

func buildCache(fsys fs.FS) (map[string]*cachedFile, error) {
	cache := make(map[string]*cachedFile)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		ct := mimeForPath(path)
		entry := &cachedFile{
			raw:         raw,
			contentType: ct,
		}
		if compressibleTypes[baseType(ct)] {
			compressed, compErr := compressBrotli(raw)
			if compErr != nil {
				return fmt.Errorf("compress %s: %w", path, compErr)
			}
			entry.compressed = compressed
		}
		urlPath := "/" + path
		cache[urlPath] = entry
		if strings.HasSuffix(path, "index.html") {
			dir := "/" + strings.TrimSuffix(path, "index.html")
			if dir == "/" {
				cache["/"] = entry
			} else {
				cache[strings.TrimSuffix(dir, "/")] = entry
				cache[dir] = entry
			}
		}
		return nil
	})
	return cache, err
}

func acceptsBrotli(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "br" || strings.HasPrefix(trimmed, "br;") {
			return true
		}
	}
	return false
}

func isCurl(r *http.Request) bool {
	return strings.HasPrefix(r.UserAgent(), "curl/")
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" && isCurl(r) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(s.curlResponse)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(s.curlResponse)
		return
	}
	entry, ok := s.cache[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", entry.contentType)
	if entry.compressed != nil && acceptsBrotli(r) {
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Content-Length", strconv.Itoa(len(entry.compressed)))
		w.Header().Set("Vary", "Accept-Encoding")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(entry.compressed)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(entry.raw)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(entry.raw)
}

func main() {
	public, err := fs.Sub(content, "public")
	if err != nil {
		log.Fatal(err)
	}
	cache, err := buildCache(public)
	if err != nil {
		log.Fatal(err)
	}
	curlTxt, err := content.ReadFile("public/curl.txt")
	if err != nil {
		log.Fatal(err)
	}
	s := &server{
		cache:        cache,
		curlResponse: curlTxt,
	}
	log.Printf("zarazaex running on :8801")
	if err := http.ListenAndServe(":8801", s); err != nil {
		log.Fatal(err)
	}
}

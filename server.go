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

const curlResponse = `
  zarazaex - developer from Minsk

  Clearnet:   https://zarazaex.xyz
  I2P:        http://zarazaex.i2p
  Yggdrasil:  http://[204:5a61:4ca1:626d:9b0e:f2d0:4207:57b6]:8801

  Projects:
    sedec (decompiler-c)                  https://github.com/zarazaex69/sedec
    s (sway-dotfiles)                     https://github.com/zarazaex69/s
    conv3n (workflow-engine)              https://github.com/zarazaex69/conv3n
    e (ime-reverse)                       https://github.com/zarazaex69/e
    zarazaex (this-site-source-code)      https://github.com/zarazaex69/zarazaex
    ro (web-tools-on-ro.zarazaex.xyz)     https://github.com/zarazaex69/ro
    l (program-pack-for-sway-dotfiles)    https://github.com/zarazaex69/l

  Contacts:
    Telegram:  @zarazaex / @zarazaexe
    GitHub:    https://github.com/zarazaex69
    Habr:      https://habr.com/ru/users/zarazaexe/
    Email:     zarazaex@tuta.io

  GPG Key:     https://zarazaex.xyz/key.txt
`

func isCurl(r *http.Request) bool {
	return strings.HasPrefix(r.UserAgent(), "curl/")
}

type curlAwareHandler struct {
	fileServer http.Handler
}

func (h *curlAwareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" && isCurl(r) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(curlResponse))
		return
	}
	h.fileServer.ServeHTTP(w, r)
}

func main() {
	public, err := fs.Sub(content, "public")
	if err != nil {
		log.Fatal(err)
	}
	handler := &curlAwareHandler{
		fileServer: http.FileServer(http.FS(public)),
	}
	http.Handle("/", handler)
	log.Println("zarazaex running on :8801")
	if err := http.ListenAndServe(":8801", nil); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed public/*
var content embed.FS

func main() {
	public, err := fs.Sub(content, "public")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(public)))
	log.Println("zarazaex running on :8801")
	if err := http.ListenAndServe(":8801", nil); err != nil {
		log.Fatal(err)
	}
}

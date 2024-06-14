package main

import (
    "log"
    "net/http"
    "sync"
)

var (
    filesDir = "./files"
    mutex    sync.Mutex
)

func main() {
    mux := http.NewServeMux()
	mux.HandleFunc("/drive/", ServeIndex)
    mux.HandleFunc("/drive/upload", UploadFile)
    mux.HandleFunc("/drive/download/", DownloadFile)
	mux.HandleFunc("/drive/create/", CreateDirectory)
    mux.HandleFunc("/drive/delete/", DeleteItem)
    mux.HandleFunc("/drive/list/", ListItems)

    log.Println("Starting server on :8080...")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}


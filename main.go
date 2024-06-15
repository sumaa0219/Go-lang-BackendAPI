package main

import (
    "log"
    "net/http"
    "sync"
    "github.com/rs/cors"
)

var (
    filesDir = "./files"
    mutex    sync.Mutex
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/drive/", ServeIndex)
    mux.HandleFunc("/drive/upload/", UploadFile)
    mux.HandleFunc("/drive/download/", DownloadFile)
    mux.HandleFunc("/drive/create/", CreateDirectory)
    mux.HandleFunc("/drive/delete/", DeleteItem)
    mux.HandleFunc("/drive/list/", ListItems)

    // CORS設定
    c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"}, // クライアントのURLを許可
        AllowedMethods: []string{"GET", "POST", "DELETE"},
        AllowedHeaders: []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    })

    handler := c.Handler(mux)

    log.Println("Starting server on :8080...")
    if err := http.ListenAndServe(":8080", handler); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
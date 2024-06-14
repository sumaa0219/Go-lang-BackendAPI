package main


import (
	"fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "mime"
    "log"
    "encoding/json"
)


func ServeIndex(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/drive/" {
        http.NotFound(w, r)
        return
    }
    http.ServeFile(w, r, "debug.html")
}



func UploadFile(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Failed to read file from request", http.StatusBadRequest)
        return
    }
    defer file.Close()

    mutex.Lock()
    defer mutex.Unlock()

    if _, err := os.Stat(filesDir); os.IsNotExist(err) {
        os.Mkdir(filesDir, os.ModePerm)
    }

    filePath := filepath.Join(filesDir, header.Filename)
    destFile, err := os.Create(filePath)
    if err != nil {
        http.Error(w, "Failed to create file", http.StatusInternalServerError)
        return
    }
    defer destFile.Close()

    if _, err := io.Copy(destFile, file); err != nil {
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "File uploaded successfully: %s\n", header.Filename)
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
    queryValues := r.URL.Query()
    fileType := queryValues.Get("type")
    fileName := strings.TrimPrefix(r.URL.Path, "/drive/download/")
    if fileName == "" {
        http.Error(w, "File name is required", http.StatusBadRequest)
        return
    }

    mutex.Lock()
    defer mutex.Unlock()

    filePath := filepath.Join(filesDir, fileName)
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    file, err := os.Open(filePath)
    if err != nil {
        http.Error(w, "Failed to open file", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    // ファイルのMIMEタイプを取得
    log.Println(fileType)   
    contentType := mime.TypeByExtension(filepath.Ext(fileName))
    if contentType == "" {
        contentType = "application/octet-stream"
    }

    // Content-Dispositionヘッダーを設定
    disposition := "attachment"
    if fileType == "inline" {
        disposition = "inline"
    }
    w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"",disposition, fileName))
    w.Header().Set("Content-Type", contentType)
    io.Copy(w, file)
}

func CreateDirectory(w http.ResponseWriter, r *http.Request) {
    dirName := strings.TrimPrefix(r.URL.Path, "/drive/create/")
    if dirName == "" {
        http.Error(w, "Directory name is required", http.StatusBadRequest)
        return
    }

    mutex.Lock()
    defer mutex.Unlock()

    dirPath := filepath.Join(filesDir, dirName)
    if err := os.Mkdir(dirPath, os.ModePerm); err != nil {
        http.Error(w, "Failed to create directory", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Directory created successfully: %s\n", dirName)
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
    itemName := strings.TrimPrefix(r.URL.Path, "/drive/delete/")
    if itemName == "" {
        http.Error(w, "Item name is required", http.StatusBadRequest)
        return
    }

    mutex.Lock()
    defer mutex.Unlock()

    itemPath := filepath.Join(filesDir, itemName)
    info, err := os.Stat(itemPath)
    if os.IsNotExist(err) {
        http.Error(w, "Item not found", http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, "Failed to get item info", http.StatusInternalServerError)
        return
    }

    if info.IsDir() {
        if err := os.RemoveAll(itemPath); err != nil {
            http.Error(w, "Failed to delete directory", http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "Directory deleted successfully: %s\n", itemName)
    } else {
        if err := os.Remove(itemPath); err != nil {
            http.Error(w, "Failed to delete file", http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "File deleted successfully: %s\n", itemName)
    }
}


func ListItems(w http.ResponseWriter, r *http.Request) {
    dirName := strings.TrimPrefix(r.URL.Path, "/drive/list/")
    dirPath := filepath.Join(filesDir, dirName)

    mutex.Lock()
    defer mutex.Unlock()

    files, err := os.ReadDir(dirPath)
    if err != nil {
        http.Error(w, "Failed to list directories and files", http.StatusInternalServerError)
        return
    }

    items := make(map[string]map[string]interface{})
    for _, file := range files {
        itemType := "file"
        if file.IsDir() {
            itemType = "dir"
        }
        
        info, err := file.Info()
        if err != nil {
            http.Error(w, "Failed to get file info", http.StatusInternalServerError)
            return
        }

        item := map[string]interface{}{
            "name":    file.Name(),
            "type":    itemType,
            "date": info.ModTime().Format("2006-01-02"),
            "time": info.ModTime().Format("15:04:05"),
        }

        if !file.IsDir() {
            item["size"] = info.Size()
        }

        items[file.Name()] = item
    }

    jsonResponse, err := json.Marshal(items)
    if err != nil {
        http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonResponse)
}





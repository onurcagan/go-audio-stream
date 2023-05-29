package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	 http.HandleFunc("/stream", mp3Handler)

    log.Println("Server listening on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func mp3Handler(w http.ResponseWriter, r *http.Request) {
    // Open the MP3 file
    file, err := os.Open("test.mp3")
    if err != nil {
        http.Error(w, "Failed to open MP3 file", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    // Set the appropriate content type for streaming MP3
    w.Header().Set("Content-Type", "audio/mpeg")

    // Copy the MP3 file data to the response writer
    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, "Failed to stream MP3", http.StatusInternalServerError)
        return
    }
}

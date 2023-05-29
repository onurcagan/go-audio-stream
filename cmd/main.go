package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

    // Get file info for content length
    fileInfo, err := file.Stat()
    if err != nil {
        http.Error(w, "Failed to get file info", http.StatusInternalServerError)
        return
    }
    fileSize := fileInfo.Size()

    // Set the appropriate content type for streaming MP3
    w.Header().Set("Content-Type", "audio/mpeg")

    // Check if range header exists
    rangeHeader := r.Header.Get("Range")
    if rangeHeader != "" {
        // Parse the range header
        ranges, err := parseRangeHeader(rangeHeader, fileSize)
        if err != nil {
            http.Error(w, "Invalid range header", http.StatusBadRequest)
            return
        }

        // Check if multiple ranges were requested (not supported)
        if len(ranges) > 1 {
            http.Error(w, "Multiple ranges not supported", http.StatusRequestedRangeNotSatisfiable)
            return
        }

        // Get the requested range
        start := ranges[0].start
        end := ranges[0].end

        // Set the appropriate status code and headers for byte range response
        w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
        w.Header().Set("Accept-Ranges", "bytes")
        w.WriteHeader(http.StatusPartialContent)

        // Seek to the start position in the file
        _, err = file.Seek(start, io.SeekStart)
        if err != nil {
            http.Error(w, "Failed to seek in file", http.StatusInternalServerError)
            return
        }

        // Create a limited reader for the requested range
        limitedReader := &io.LimitedReader{R: file, N: end - start + 1}

        // Copy the requested range to the response writer
        _, err = io.Copy(w, limitedReader)
        if err != nil {
            http.Error(w, "Failed to stream MP3", http.StatusInternalServerError)
            return
        }
    } else {
        // No range requested, serve the full file
        w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
        _, err = io.Copy(w, file)
        if err != nil {
            http.Error(w, "Failed to stream MP3", http.StatusInternalServerError)
            return
        }
    }
}

type byteRange struct {
    start int64
    end   int64
}

func parseRangeHeader(rangeHeader string, fileSize int64) ([]byteRange, error) {
    var ranges []byteRange

    rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
    rangeParts := strings.Split(rangeStr, ",")
    for _, part := range rangeParts {
        rangeSpec := strings.Split(part, "-")
        if len(rangeSpec) != 2 {
            return nil, errors.New("Invalid range specification")
        }

        start, err := strconv.ParseInt(rangeSpec[0], 10, 64)
        if err != nil {
            return nil, err
        }

        var end int64
        if rangeSpec[1] == "" {
            end = fileSize - 1
        } else {
            end, err = strconv.ParseInt(rangeSpec[1], 10, 64)
            if err != nil {
                return nil, err
            }
        }

        if start > end || end >= fileSize {
            return nil, errors.New("Invalid range values")
        }

        ranges = append(ranges, byteRange{start: start, end: end})
    }

    return ranges, nil
}
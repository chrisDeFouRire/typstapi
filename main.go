package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/typst/", handleTypst)
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleTypst(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the filename from the URL path
	filename := filepath.Base(r.URL.Path)
	if filename == "" || filename == "typst" {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "typst-*")
	if err != nil {
		http.Error(w, "Failed to create temporary directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir) // Clean up when done

	// Parse the multipart form
	err = r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Save uploaded files
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Failed to open uploaded file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Create the file in the temp directory
			dst, err := os.Create(filepath.Join(tempDir, fileHeader.Filename))
			if err != nil {
				http.Error(w, "Failed to create file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			if _, err = io.Copy(dst, file); err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
		}
	}

	// Get JSON data and save to data.json
	jsonData := r.FormValue("data")
	if jsonData != "" {
		// Verify it's valid JSON
		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(jsonData), &jsonMap); err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}

		// Write to data.json
		if err := os.WriteFile(filepath.Join(tempDir, "data.json"), []byte(jsonData), 0644); err != nil {
			http.Error(w, "Failed to save JSON data", http.StatusInternalServerError)
			return
		}
	}

	// Run typst command with stderr capture
	cmd := exec.Command("typst", "compile", filename)
	cmd.Dir = tempDir

	// Capture stderr
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Include the stderr output in the error message
		errMsg := fmt.Sprintf("Failed to compile typst document: %v\n\nTypst Error Output:\n%s",
			err, stderr.String())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Read the generated PDF
	pdfPath := filepath.Join(tempDir, filepath.Base(filename[:len(filename)-4]+".pdf"))
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		http.Error(w, "Failed to read generated PDF", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(pdfPath)))

	// Create gzip writer
	gz := gzip.NewWriter(w)
	defer gz.Close()

	// Write the PDF data
	if _, err := gz.Write(pdfData); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
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

// getPDFFiles returns a sorted list of PDF files from the multipart form that match the prefix
func getPDFFiles(form *multipart.Form, prefix string) []string {
	var files []string
	for filename := range form.File {
		if strings.HasPrefix(filename, prefix) && strings.HasSuffix(strings.ToLower(filename), ".pdf") {
			files = append(files, filename)
		}
	}
	sort.Strings(files)
	return files
}

func handleTypst(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := filepath.Base(r.URL.Path)
	if filename == "" || filename == "typst" {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	tempDir, err := os.MkdirTemp("", "typst-*")
	if err != nil {
		http.Error(w, "Failed to create temporary directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	if err := saveFormFiles(r, tempDir); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	typstPDFPath, err := compileTypst(tempDir, filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdfData, err := mergePDFs(tempDir, typstPDFPath, r.MultipartForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendPDFResponse(w, r, pdfData)
}

func saveFormFiles(r *http.Request, tempDir string) error {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return fmt.Errorf("Failed to parse form: %w", err)
	}

	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				return fmt.Errorf("Failed to open uploaded file: %w", err)
			}
			defer file.Close()

			dst, err := os.Create(filepath.Join(tempDir, fileHeader.Filename))
			if err != nil {
				return fmt.Errorf("Failed to create file: %w", err)
			}
			defer dst.Close()

			if _, err = io.Copy(dst, file); err != nil {
				return fmt.Errorf("Failed to save file: %w", err)
			}
		}
	}

	jsonData := r.FormValue("data")
	if jsonData != "" {
		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(jsonData), &jsonMap); err != nil {
			return fmt.Errorf("Invalid JSON data: %w", err)
		}

		if err := os.WriteFile(filepath.Join(tempDir, "data.json"), []byte(jsonData), 0644); err != nil {
			return fmt.Errorf("Failed to save JSON data: %w", err)
		}
	}
	return nil
}

func compileTypst(dir, filename string) (string, error) {
	cmd := exec.Command("typst", "compile", filename)
	cmd.Dir = dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Failed to compile typst document: %v\n\nTypst Error Output:\n%s", err, stderr.String())
	}

	return filepath.Join(dir, filepath.Base(filename[:len(filename)-4]+".pdf")), nil
}

func mergePDFs(dir, typstPDFPath string, form *multipart.Form) ([]byte, error) {
	prePDFs := getPDFFiles(form, "pre_")
	postPDFs := getPDFFiles(form, "post_")

	if len(prePDFs) == 0 && len(postPDFs) == 0 {
		return os.ReadFile(typstPDFPath)
	}

	var pdfsToMerge []string
	for _, name := range prePDFs {
		pdfsToMerge = append(pdfsToMerge, filepath.Join(dir, name))
	}
	pdfsToMerge = append(pdfsToMerge, typstPDFPath)
	for _, name := range postPDFs {
		pdfsToMerge = append(pdfsToMerge, filepath.Join(dir, name))
	}

	mergedPDFPath := filepath.Join(dir, "merged.pdf")
	if err := api.MergeAppendFile(pdfsToMerge, mergedPDFPath, false, nil); err != nil {
		return nil, fmt.Errorf("Failed to merge PDFs: %w", err)
	}

	return os.ReadFile(mergedPDFPath)
}

func sendPDFResponse(w http.ResponseWriter, r *http.Request, pdfData []byte) {
	// Set response headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=output.pdf")

	// Check if client accepts gzip encoding
	acceptEncoding := r.Header.Get("Accept-Encoding")
	supportsGzip := strings.Contains(acceptEncoding, "gzip")

	// Use gzip compression only if client supports it
	if supportsGzip {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		if _, err := gz.Write(pdfData); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	} else {
		// Send uncompressed data
		if _, err := w.Write(pdfData); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

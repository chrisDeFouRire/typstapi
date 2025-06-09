package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// helper to create a multipart.Form with the given file names
func createForm(filenames []string) *multipart.Form {
	files := make(map[string][]*multipart.FileHeader)
	for _, name := range filenames {
		files[name] = []*multipart.FileHeader{{Filename: name}}
	}
	return &multipart.Form{File: files}
}

func TestGetPDFFilesFiltering(t *testing.T) {
	form := createForm([]string{
		"pre_b.pdf",
		"pre_a.pdf",
		"post_c.pdf",
		"post_b.pdf",
		"post_a.pdf",
		"pre_foo.txt",
		"random.pdf",
		"post_d.doc",
		"pre_c.PDF",
	})

	gotPre := getPDFFiles(form, "pre_")
	wantPre := []string{"pre_a.pdf", "pre_b.pdf", "pre_c.PDF"}
	if !reflect.DeepEqual(gotPre, wantPre) {
		t.Errorf("pre_ files mismatch:\nwant %v\n got %v", wantPre, gotPre)
	}

	gotPost := getPDFFiles(form, "post_")
	wantPost := []string{"post_a.pdf", "post_b.pdf", "post_c.pdf"}
	if !reflect.DeepEqual(gotPost, wantPost) {
		t.Errorf("post_ files mismatch:\nwant %v\n got %v", wantPost, gotPost)
	}
}

func TestGetPDFFilesSorted(t *testing.T) {
	form := createForm([]string{"pre_c.pdf", "pre_a.pdf", "pre_b.pdf"})
	got := getPDFFiles(form, "pre_")
	want := []string{"pre_a.pdf", "pre_b.pdf", "pre_c.pdf"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected sorted result %v, got %v", want, got)
	}
}

// Test saving uploaded files and JSON data
func TestSaveFormFiles(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// create a simple file field
	fw, err := writer.CreateFormFile("foo.txt", "foo.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write([]byte("hello")); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// add JSON data
	if err := writer.WriteField("data", `{"a":1}`); err != nil {
		t.Fatalf("write field: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	tempDir := t.TempDir()

	if err := saveFormFiles(req, tempDir); err != nil {
		t.Fatalf("saveFormFiles failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tempDir, "foo.txt")); err != nil {
		t.Fatalf("expected file saved: %v", err)
	}

	dataPath := filepath.Join(tempDir, "data.json")
	b, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("expected data.json: %v", err)
	}
	if string(b) != `{"a":1}` {
		t.Fatalf("unexpected json contents: %s", string(b))
	}
}

// Test compiling a typst document
func TestCompileTypst(t *testing.T) {
	tmp := t.TempDir()

	// copy example files
	for _, name := range []string{"main.typ", "data.json", "splash192.png"} {
		src := filepath.Join("example", name)
		dst := filepath.Join(tmp, name)
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	pdfPath, err := compileTypst(tmp, "main.typ")
	if err != nil {
		t.Fatalf("compileTypst failed: %v", err)
	}
	info, err := os.Stat(pdfPath)
	if err != nil {
		t.Fatalf("compiled pdf missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("compiled pdf is empty")
	}
}

// Test merging PDFs with pre and post files
func TestMergePDFs(t *testing.T) {
	tmp := t.TempDir()

	// create two simple typst PDFs
	for i := 1; i <= 2; i++ {
		typFile := filepath.Join(tmp, fmt.Sprintf("file%d.typ", i))
		content := []byte("= Test\nHello")
		os.WriteFile(typFile, content, 0644)
		if _, err := compileTypst(tmp, filepath.Base(typFile)); err != nil {
			t.Fatalf("compile %s: %v", typFile, err)
		}
		pdfName := fmt.Sprintf("pre_%d.pdf", i)
		if i == 2 {
			pdfName = "post_1.pdf"
		}
		os.Rename(strings.TrimSuffix(typFile, ".typ")+".pdf", filepath.Join(tmp, pdfName))
	}

	// typst output
	// copy example files required for compilation
	for _, name := range []string{"main.typ", "data.json", "splash192.png"} {
		b, _ := os.ReadFile(filepath.Join("example", name))
		os.WriteFile(filepath.Join(tmp, name), b, 0644)
	}
	pdfPath, err := compileTypst(tmp, "main.typ")
	if err != nil {
		t.Fatalf("compile example: %v", err)
	}

	form := createForm([]string{"pre_1.pdf", "post_1.pdf"})
	merged, err := mergePDFs(tmp, pdfPath, form)
	if err != nil {
		t.Fatalf("mergePDFs failed: %v", err)
	}
	if len(merged) == 0 {
		t.Fatal("merged pdf empty")
	}
}

// Integration test using example files
func TestHandleTypstIntegration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handleTypst))
	defer srv.Close()

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	for _, name := range []string{"main.typ", "data.json", "splash192.png"} {
		fw, err := writer.CreateFormFile(name, name)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		data, err := os.ReadFile(filepath.Join("example", name))
		if err != nil {
			t.Fatalf("read example %s: %v", name, err)
		}
		fw.Write(data)
	}
	writer.Close()

	resp, err := http.Post(srv.URL+"/typst/main.typ", writer.FormDataContentType(), buf)
	if err != nil {
		t.Fatalf("post request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "pdf") {
		t.Fatalf("expected pdf content-type, got %s", ct)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty pdf data")
	}
}

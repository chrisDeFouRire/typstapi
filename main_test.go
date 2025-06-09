package main

import (
	"mime/multipart"
	"reflect"
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

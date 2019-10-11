package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const (
	maxUploadSizeBytes = 10 * 1024 // 10Mb
	uploadPathM3204    = "./m3204"
	uploadPathM3205    = "./m3205"
	maxFileNameLength  = 20

	port = ":8080"
)

func renderError(w http.ResponseWriter, msg string, status int) {
	fmt.Println(">>>> " + msg)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

func uploadFileHandler(uploadPath string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSizeBytes)
		if r.ContentLength > maxUploadSizeBytes {
			renderError(w, "the file is too big ", http.StatusBadRequest)
			return
		}

		fileType := ".tar"

		file := r.Body
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID FILE 1", http.StatusBadRequest)
			return
		}

		fileName := r.Header.Get("student_name")
		var validName = regexp.MustCompile(`^[a-zA-Z]+$`)
		if !validName.MatchString(fileName) && len(fileName) < maxFileNameLength {
			renderError(w, "Wrong student_name."+fileName+" Should be in one word, ex. IvanIvanov",
				http.StatusBadRequest)
			return
		}

		newPath := filepath.Join(uploadPath, fileName+fileType)
		fmt.Printf("FileType: %s, File: %s\n", fileType, newPath)

		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "can't create a new file on FS "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer newFile.Close()

		if _, err := newFile.Write(fileBytes); err != nil {
			renderError(w, "can't wright file to FS", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("success"))
	})
}

func main() {
	http.HandleFunc("/upload/m3204", uploadFileHandler(uploadPathM3204))
	http.HandleFunc("/upload/m3205", uploadFileHandler(uploadPathM3205))

	fmt.Println("Server started on default port", port)
	fmt.Println(http.ListenAndServe(port, nil))
}
package main

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const (
	maxUploadSizeBytes = 10 * 1024 * 1024 // 10Mb
	uploadPathM3204    = "./m3204/lab"
	uploadPathM3205    = "./m3205/lab"
	maxFileNameLength  = 20

	port = ":8080"
)

const fileType = ".tar"

func renderError(w http.ResponseWriter, msg string, status int) {
	fmt.Printf(">>>> [%d] : %s\n", time.Now().Unix(), msg)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

func uploadFileHandler(uploadPath string, labNum int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSizeBytes)
		if r.ContentLength > maxUploadSizeBytes {
			renderError(w, "the file is too big ", http.StatusBadRequest)
			return
		}

		err := r.ParseMultipartForm(maxUploadSizeBytes)
		file, _, err := r.FormFile("data")
		if file == nil {
			renderError(w, "wrong data field: check @ and filename", http.StatusBadRequest)
			return
		}

		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "can't read the file", http.StatusBadRequest)
			return
		}

		fileName := r.Header.Get("student_name")
		var validName = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
		if !validName.MatchString(fileName) && len(fileName) < maxFileNameLength {
			renderError(
				w,
				"Wrong student_name. "+fileName+" Should be in one word, ex. IvanIvanov, nums are possible",
				http.StatusBadRequest,
			)
			return
		}

		newPath := filepath.Join(uploadPath+strconv.Itoa(labNum), fileName+fileType)
		fmt.Printf("[%d]: saving FileType: %s, File: %s\n", time.Now().Unix(), fileType, newPath)

		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "can't create a new file on FS: "+err.Error(), http.StatusInternalServerError)
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

func checkValidTar(name string) bool {
	file, err := os.Open(name)
	if err != nil {
		return false
	}
	defer file.Close()
	_, err = tar.NewReader(file).Next()
	if err != nil {
		return false
	}
	return true
}

func infoHandler(group int8) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%d] Sending info m320%d\n", time.Now().Unix(), group)
		path := uploadPathM3204
		if group == 5 {
			path = uploadPathM3205
		}

		var tarFile = regexp.MustCompile(`^*.tar$`)

		info := make(map[int8][]string)
		for i := int8(1); i < 7; i++ {
			fmt.Println("Reading: " + path + strconv.Itoa(int(i)))
			files, err := ioutil.ReadDir(path + strconv.Itoa(int(i)))
			if err != nil {
				renderError(w, "it's not possible to get files", http.StatusInternalServerError)
				continue
			}
			filesSlice := make([]string, 0, len(files))
			for j := range files {
				fmt.Printf("Match str: %s %t\n", files[j].Name(), tarFile.MatchString(files[j].Name()))
				if files[j].IsDir() && tarFile.MatchString(files[j].Name()) {
					comment := ""
					if checkValidTar(path + strconv.Itoa(int(i)) + files[j].Name()) {
						comment = " :the tar is not valid, need to be reloaded"
					}
					filesSlice = append(filesSlice, files[j].Name()+comment)
				}
			}
			info[i] = filesSlice
		}

		for k, v := range info {
			fmt.Println(k)
			fmt.Println(v)
		}

		data, _ := json.Marshal(info)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
}

func main() {
	for i := 1; i < 7; i++ {
		http.HandleFunc("/upload/m3204/lab"+strconv.Itoa(i), uploadFileHandler(uploadPathM3204, i))
		http.HandleFunc("/upload/m3205/lab"+strconv.Itoa(i), uploadFileHandler(uploadPathM3205, i))
	}

	http.HandleFunc("/info/m3204", infoHandler(4))
	http.HandleFunc("/info/m3205", infoHandler(5))

	fmt.Println("Server started on default port", port)
	fmt.Println(http.ListenAndServe(port, nil))
}

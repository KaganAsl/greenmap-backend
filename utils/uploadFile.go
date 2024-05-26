package utils

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func UploadFile(r *http.Request) (string, error) {

	// It supports IMAGES

	err := r.ParseMultipartForm(10 << 20) // limit your maxMultipartMemory
	if err != nil {
		return "", err
	}

	file, header, err := r.FormFile("File") // "File" is the form field name
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	parentDir := filepath.Dir(dir) // Get the parent dir for ../uploads folder

	err = os.MkdirAll(filepath.Join(parentDir, "uploads"), 0755)
	if err != nil {
		return "", err
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		log.Println("Error Reading the File", err)
		return "", err
	}

	// Reset the read pointer to start of the file after reading
	file.Seek(0, 0)

	mimeType := http.DetectContentType(buffer)

	if !strings.HasPrefix(mimeType, "image/") {
		log.Println("The uploaded file is not an image")
		return "", err
	}

	filenameSplitted := strings.Split(header.Filename, ".")
	fileExtension := filenameSplitted[len(filenameSplitted)-1]

	filename = Base64EncodeString(CreateToken(header.Filename, fileExtension, time.Now().String()))
	filename = filename + "." + fileExtension

	dst, err := os.Create(filepath.Join(parentDir, "uploads", filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filename, nil
}

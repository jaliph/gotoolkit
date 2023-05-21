package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_+"

// Tools is type used to intantiate this module, Any variable if this type will have access to all
// the methods with the reciever *Tools
type Tools struct {
	MaxAllowedSize   int64
	AllowedFileTypes []string
}

// RandomString returns a string of random characters if length n, using randomStringSource as the source
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}

type UploadFile struct {
	NewFileName  string
	OriginalName string
	FileSize     int64
}

func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadFile, error) {
	renameFlag := false
	if len(rename) > 0 {
		renameFlag = true
	}

	files, err := t.UploadFiles(r, uploadDir, renameFlag)
	if len(files) == 1 {
		return files[0], err
	} else {
		return nil, fmt.Errorf("run error: %w", err)
	}

}

// takes a http request and saves the file from the multipart
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadFile, error) {
	renameFlag := false
	if len(rename) > 0 {
		renameFlag = rename[0]
	}

	var uploadedFiles []*UploadFile

	if t.MaxAllowedSize == 0 {
		t.MaxAllowedSize = 1024 * 1034 * 124
	}

	err := r.ParseMultipartForm(int64(t.MaxAllowedSize))
	if err != nil {
		return nil, errors.New("the uploaded file is too big")
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadFile) ([]*UploadFile, error) {
				var uploadedFile UploadFile

				inFile, err := hdr.Open()
				if err != nil {
					return nil, err
				}

				defer inFile.Close()

				buffer := make([]byte, 512)

				_, err = inFile.Read(buffer)
				if err != nil {
					return nil, err
				}

				allowed := false
				fileType := http.DetectContentType(buffer)

				// allowedTypes := []string{"image/jpeg, image/gif", "image/png"}

				if len(t.AllowedFileTypes) > 0 {
					for _, allowedtype := range t.AllowedFileTypes {
						if strings.EqualFold(fileType, allowedtype) {
							allowed = true
							break
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("file type is not allowed")
				}

				_, err = inFile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				uploadedFile.OriginalName = hdr.Filename
				if renameFlag {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s\n", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				var outFile *os.File

				defer outFile.Close()

				if outFile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err == nil {
					fileSize, err := io.Copy(outFile, inFile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				} else {
					return nil, err
				}
				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)

			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, err
}

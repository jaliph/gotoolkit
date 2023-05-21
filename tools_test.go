package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("Wrong length random string generated.")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	// {name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false},
	// {name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false},
	{name: "not allowed", allowedTypes: []string{"image/jpeg"}, renameFile: false, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, test := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testData/img.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testData/img.png")
			if err != nil {

				t.Error(err)
			}

			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("Error decoding the image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools

		testTools.AllowedFileTypes = test.allowedTypes

		uploadFiles, err := testTools.UploadFiles(request, "./testData/uploads", test.renameFile)
		if err != nil && !test.errorExpected {
			t.Error(err)
		}

		if !test.errorExpected {
			if _, err = os.Stat(fmt.Sprintf("./testData/uploads/%s", uploadFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: Should have got a file : %s", test.name, err.Error())
			}
			_ = os.Remove(fmt.Sprintf("./testData/uploads/%s", uploadFiles[0].NewFileName))
		}

		if !test.errorExpected && err != nil {
			t.Errorf("%s: Error was not expected but still found : %s", test.name, err.Error())
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	for _, test := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		go func() {
			defer writer.Close()

			part, err := writer.CreateFormFile("file", "./testData/img.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testData/img.png")
			if err != nil {

				t.Error(err)
			}

			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("Error decoding the image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools

		testTools.AllowedFileTypes = test.allowedTypes
		uploadFile, err := testTools.UploadOneFile(request, "./testData/uploads", test.renameFile)
		if err != nil && !test.errorExpected {
			t.Error(err)
		}

		if !test.errorExpected {
			if _, err = os.Stat(fmt.Sprintf("./testData/uploads/%s", uploadFile.NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: Should have got a file : %s", test.name, err.Error())
			}
			_ = os.Remove(fmt.Sprintf("./testData/uploads/%s", uploadFile.NewFileName))
		}

		if !test.errorExpected && err != nil {
			t.Errorf("%s: Error was not expected but still found : %s", test.name, err.Error())
		}
	}
}

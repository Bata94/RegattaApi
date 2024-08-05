package utils

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SavePDFfromHTML(htmlUrl, subDir, filename string, footer bool) (string, error) {
	// Prepare a buffer to write the request body
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Add the URL parameter
	multipartWriter.WriteField("url", "http://api-dev:8080/api/v1/"+htmlUrl)
	multipartWriter.WriteField("paperWidth", "8.27")
	multipartWriter.WriteField("paperHeight", "11.69")

	if footer {
		// Get Footerfile
		// TODO: unnecessary HTTP Request
		footerReq, footerReqError := http.Get("http://localhost:8080/api/v1/leitung/pdfFooter")
		if footerReqError != nil {
			return "", footerReqError
		}
		defer footerReq.Body.Close()
		footerContent, _ := io.ReadAll(footerReq.Body)

		// Add the footer file to the multipart form data
		footerWriter, err := multipartWriter.CreateFormFile("files", "footer.html")
		if err != nil {
			return "", err
		}
		_, err = footerWriter.Write(footerContent)
		if err != nil {
			return "", err
		}
	}

	multipartWriter.Close()

	req, err := http.NewRequest("POST", "http://gotenberg:3000/forms/chromium/convert/url", &requestBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	} else if resp.StatusCode != 200 {
		return "", errors.New("gotenberg error: " + resp.Status)
	}
	defer resp.Body.Close()

	if !strings.HasSuffix(filename, ".pdf") {
		filename += ".pdf"
	}
	basePath := filepath.Join("./files", subDir)
	err = os.MkdirAll(basePath, 0o666)
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(basePath, filename)
	outputFile, err := os.Create(filePath)
	defer outputFile.Close()

	_, err = io.Copy(outputFile, resp.Body)

	if err != nil {
		return "", err
	}

	return filePath, nil
}

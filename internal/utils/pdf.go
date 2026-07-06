package utils

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bata94/RegattaApi/internal/config"
)

const (
	writeFilePerms os.FileMode = 0o666
	paperWidth                  = "8.27"
	paperHeight                 = "11.69"
	pdfSuffix                   = ".pdf"
)

func SavePDFfromHTML(htmlUrl, subDir, filename string, footer bool) (string, error) {
	// Prepare a buffer to write the request body
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Add the URL parameter
	multipartWriter.WriteField("url", config.C.AppInternalURL+"/api/v1/"+htmlUrl)
	multipartWriter.WriteField("paperWidth", paperWidth)
	multipartWriter.WriteField("paperHeight", paperHeight)

	if footer {
		// Get Footerfile
		// TODO: unnecessary HTTP Request
		footerReq, footerReqError := http.Get(config.C.AppInternalURL + "/api/v1/leitung/pdfFooter")
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

	req, err := http.NewRequest("POST", config.C.GotenbergURL+"/forms/chromium/convert/url", &requestBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error in HttpRequest module")
		return "", err
	} else if resp.StatusCode != 200 {
		slog.Error("Error in HttpRequest Status")
		return "", errors.New("gotenberg error: " + resp.Status)
	}
	defer resp.Body.Close()

	if !strings.HasSuffix(filename, pdfSuffix) {
		filename += pdfSuffix
	}
	basePath := filepath.Join(config.C.Paths.FilesDir, subDir)
	err = os.MkdirAll(basePath, writeFilePerms)
	if err != nil {
		slog.Error("Error in Dir creation")
		return "", err
	}
	filePath := filepath.Join(basePath, filename)
	outputFile, err := os.Create(filePath)
	defer outputFile.Close()
	if err != nil {
		slog.Error("Error in creation of file")
		return "", err
	}

	_, err = io.Copy(outputFile, resp.Body)

	if err != nil {
		slog.Error("Error in Writing to file")
		return "", err
	}

	return filePath, nil
}

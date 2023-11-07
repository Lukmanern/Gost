package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/Lukmanern/gost/internal/env"
)

type UploadFile interface {
	Upload(fileHeader *multipart.FileHeader) (file_url string, err error)
	Delete(link string) (err error)
}

type client struct {
	httpClient *http.Client
	token      string
	urlProject string
	fileUrl    string
	bucketName string
}

func NewClient() UploadFile {
	config := env.Configuration()
	fileUrl := config.BucketURL + "/storage/v1/object/public/" + config.BucketName + "/"
	return &client{
		httpClient: &http.Client{},
		token:      config.BucketToken,
		fileUrl:    fileUrl,
		urlProject: config.BucketURL,
		bucketName: config.BucketName,
	}
}

func (c *client) Upload(fileHeader *multipart.FileHeader) (fileURL string, err error) {
	// Open the file associated with the file header
	file, openHeaderErr := fileHeader.Open()
	if openHeaderErr != nil {
		return "", openHeaderErr
	}
	defer file.Close()

	// Create a new multipart writer
	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	fileField, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(fileField, file)
	if copyErr != nil {
		return "", copyErr
	}
	writer.Close()
	url := c.urlProject + "/storage/v1/object/" + c.bucketName + "/" + fileHeader.Filename
	request, err := http.NewRequest(http.MethodPost, url, requestBody)
	if err != nil {
		return "", err
	}

	// Set the Bearer token in the Authorization header
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Println("Upload failed. Status code:", response.StatusCode)
		errorMessage, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return "", readErr
		}
		var errorResponse struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if unmarshalErr := json.Unmarshal(errorMessage, &errorResponse); unmarshalErr != nil {
			return "", unmarshalErr
		}
		log.Println("Error message:", errorResponse.Message+", "+errorResponse.Error)
		return "", errors.New(errorResponse.Message + ", " + errorResponse.Error)
	}

	link := c.fileUrl + fileHeader.Filename
	reqImage, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return "", err
	}
	resImage, err := c.httpClient.Do(reqImage)
	if err != nil {
		return "", err
	}
	if resImage.StatusCode != http.StatusOK {
		message := "Successfully uploaded, but failed to test-get the file. Got status code: "
		message += string(rune(response.StatusCode))
		return "", errors.New(message)
	}
	return link, nil
}

func (c *client) Delete(link string) (err error) {
	return
}

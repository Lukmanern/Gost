package service

import (
	"bytes"
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
	token      string
	httpClient *http.Client
	urlProject string
	fileUrl    string
	bucketName string
}

func NewClient() UploadFile {
	config := env.Configuration()
	fileUrl := config.StorageURL + "/storage/v1/object/public/" + config.BucketName //+ "/"
	return &client{
		httpClient: &http.Client{},
		fileUrl:    fileUrl,
		urlProject: config.StorageURL,
		bucketName: config.BucketName,
	}
}

func (c *client) Upload(fileHeader *multipart.FileHeader) (file_url string, err error) {
	file, openHeaderErr := fileHeader.Open()
	if openHeaderErr != nil {
		return "", openHeaderErr
	}
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)
	fileWriter, createFormErr := multipartWriter.CreateFormFile("file", fileHeader.Filename)
	if createFormErr != nil {
		return "", createFormErr
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return "", err
	}
	err = multipartWriter.Close()
	if err != nil {
		return "", err
	}
	url := "https://xx.supabase.co/storage/v1/object/" + c.bucketName + "/" + fileHeader.Filename
	request, err := http.NewRequest(http.MethodPost, url, &requestBody)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Println("got -1: ", response.StatusCode)
		return "", errors.New("failed upload file")
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
		return "", errors.New("success upload, failed to test-get file back, got " + string(rune(response.StatusCode)))
	}
	return link, nil
}

func (c *client) Delete(link string) (err error) {
	return
}

package service

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/gofiber/fiber/v2"
)

type UploadFile interface {
	Upload(fileHeader *multipart.FileHeader) (file_url string, err error)
	Delete(link string) (err error)
}

type client struct {
	PublicURL  string
	UploadURL  string
	Token      string
	BucketURL  string
	BucketName string
}

func NewClient() UploadFile {
	config := env.Configuration()
	return &client{
		PublicURL:  config.BucketURL + "/storage/v1/object/public/" + config.BucketName + "/",
		UploadURL:  config.BucketURL + "/storage/v1/object/" + config.BucketName + "/",
		Token:      config.BucketToken,
		BucketURL:  config.BucketURL,
		BucketName: config.BucketName,
	}
}

func (c *client) Upload(fileHeader *multipart.FileHeader) (fileURL string, err error) {
	fileName := fileHeader.Filename
	file, headerErr := fileHeader.Open()
	if headerErr != nil {
		return "", headerErr
	}
	defer file.Close()

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	fileField, formErr := writer.CreateFormFile("file", fileName)
	if formErr != nil {
		return "", formErr
	}
	_, copyErr := io.Copy(fileField, file)
	if copyErr != nil {
		return "", copyErr
	}
	writer.Close()
	url := c.UploadURL + fileName
	request, newReqErr := http.NewRequest(http.MethodPost, url, requestBody)
	if newReqErr != nil {
		return "", newReqErr
	}

	request.Header.Set("Authorization", "Bearer "+c.Token)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	response, doErr := client.Do(request)
	if doErr != nil {
		return "", doErr
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", responseErrHandler(response)
	}

	link := c.PublicURL + fileName
	reqImage, reqErr := http.NewRequest(http.MethodGet, link, nil)
	if reqErr != nil {
		return "", reqErr
	}
	resp, doErr := client.Do(reqImage)
	if doErr != nil {
		return "", doErr
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", responseErrHandler(resp)
	}
	return link, nil
}

func (c *client) Delete(link string) (err error) {
	return
}

func responseErrHandler(resp *http.Response) (err error) {
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	var errResp struct {
		Message    string `json:"message"`
		Error      string `json:"error"`
		StatusCode string `json:"statusCode"`
	}
	if unmarshalErr := json.Unmarshal(respBody, &errResp); unmarshalErr != nil {
		return unmarshalErr
	}
	statusCode, convErr := strconv.Atoi(errResp.StatusCode)
	if convErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed conv:"+convErr.Error())
	}
	return fiber.NewError(statusCode, errResp.Message+", "+errResp.Error)
}

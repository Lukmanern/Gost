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

type SortBy struct {
	Column string `json:"column"`
	Order  string `json:"order"`
}
type ListReqBody struct {
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
	SortByOptions SortBy `json:"sortBy"`
	Prefix        string `json:"prefix"`
}
type ListReponse struct {
	Name           string `json:"name"`
	ID             string `json:"id"`
	UpdatedAt      string `json:"updated_at"`
	CreatedAt      string `json:"created_at"`
	LastAccessedAt string `json:"last_accessed_at"`
	Metadata       struct {
		ETag           string `json:"eTag"`
		Size           int64  `json:"size"`
		MimeType       string `json:"mimetype"`
		CacheControl   string `json:"cacheControl"`
		LastModified   string `json:"lastModified"`
		ContentLength  int64  `json:"contentLength"`
		HttpStatusCode int64  `json:"httpStatusCode"`
	} `json:"metadata"`
}

type UploadFile interface {
	Upload(fileHeader *multipart.FileHeader) (file_url string, err error)
	Delete(link string) (err error)
	GetFilesList() (files []map[string]any, err error)
}

type client struct {
	PublicURL    string
	ListFilesURL string
	UploadURL    string
	Token        string
	BucketURL    string
	BucketName   string
}

func NewClient() UploadFile {
	config := env.Configuration()
	baseURL := config.BucketURL + "/storage/v1/object/"
	return &client{
		PublicURL:    baseURL + "public/" + config.BucketName + "/",
		ListFilesURL: baseURL + "list/" + config.BucketName,
		UploadURL:    baseURL + config.BucketName + "/",
		Token:        config.BucketToken,
		BucketURL:    config.BucketURL,
		BucketName:   config.BucketName,
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
	respUpload, doErr := client.Do(request)
	if doErr != nil {
		return "", doErr
	}
	defer respUpload.Body.Close()

	if respUpload.StatusCode != http.StatusOK {
		return "", responseErrHandler(respUpload)
	}

	link := c.PublicURL + fileName
	reqTestGet, reqErr := http.NewRequest(http.MethodGet, link, nil)
	if reqErr != nil {
		return "", reqErr
	}
	respTestGet, doErr := client.Do(reqTestGet)
	if doErr != nil {
		return "", doErr
	}
	defer respTestGet.Body.Close()
	if respTestGet.StatusCode != http.StatusOK {
		return "", responseErrHandler(respTestGet)
	}
	return link, nil
}

func (c *client) Delete(link string) (err error) {
	return
}

func (c *client) GetFilesList() (files []map[string]any, err error) {
	body := ListReqBody{
		Limit:  999,
		Offset: 1,
		Prefix: "",
		SortByOptions: SortBy{
			Column: "name",
			Order:  "asc",
		},
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	listFileURL := c.ListFilesURL
	request, err := http.NewRequest(http.MethodPost, listFileURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+c.Token)
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, responseErrHandler(response)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, responseErrHandler(response)
	}

	respBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}
	var listResp []ListReponse

	if unmarshalErr := json.Unmarshal(respBody, &listResp); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	for _, list := range listResp {
		file := map[string]any{
			"name":        list.Name,
			"uploaded_at": list.CreatedAt,
			"size":        list.Metadata.ContentLength,
		}
		files = append(files, file)
	}
	return files, nil
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
		return fiber.NewError(fiber.StatusInternalServerError, "failed conv: "+convErr.Error())
	}
	return fiber.NewError(statusCode, errResp.Message+", "+errResp.Error)
}

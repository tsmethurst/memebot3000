package mastodon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	authEndpoint   = "/oauth/authorize"
	tokenEndpoint  = "/oauth/token"
	mediaEndpoint  = "/api/v1/media"
	statusEndpoint = "/api/v1/statuses"
)

// Client represents a client that can interact with mastodon
type Client interface {
	PublishMedia(filename string) (string, error)
	PublishStatus(status string, mediaID string, sensitive bool, spoilerText string) error
}

// client implements Client
type client struct {
	mastodonURL  string
	accessToken  string
}

// NewClient returns a new mastodon client ready to go
func NewClient(mastodonURL string, mastodonAccessToken string) Client {
	return &client{
		mastodonURL:  mastodonURL,
		accessToken:  mastodonAccessToken,
	}
}

func (c *client) PublishMedia(filename string) (string, error) {
	buffer, filetype, err := createMultipartFormData("file", filename)
	if err != nil {
		return "", fmt.Errorf("error creating multipart form data: %s", err)
	}
	reqURL := fmt.Sprintf("%s%s", c.mastodonURL, mediaEndpoint)
	req, err := http.NewRequest(http.MethodPost, reqURL, buffer)
	req.Header.Set("Content-Type", filetype)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing post to %s: %s", reqURL, err)
	}
	defer resp.Body.Close()

	mediaBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading media response body: %s", err)
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return "", fmt.Errorf("received code %d from %s with response body: %s", resp.StatusCode, reqURL, mediaBody)
	}

	mr := &MediaResponse{}
	if err := json.Unmarshal(mediaBody, mr); err != nil {
		return "", fmt.Errorf("error parsing media response body: %s", err)
	}

	if mr.MediaID == "" {
		return "", errors.New("no media id returned")
	}

	return mr.MediaID, nil
}

func (c *client) PublishStatus(status string, mediaID string, sensitive bool, spoilerText string) error {
	reqURL := fmt.Sprintf("%s%s", c.mastodonURL, statusEndpoint)
	form := url.Values{
		"status":       []string{status},
		"media_ids[]":  []string{mediaID},
		"sensitive":    []string{fmt.Sprintf("%t", sensitive)},
		"spoiler_text": []string{spoilerText},
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error doing post to %s: %s", reqURL, err)
	}
	defer resp.Body.Close()

	statusBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading status response body: %s", err)
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return fmt.Errorf("received code %d from %s with response body: %s", resp.StatusCode, reqURL, statusBody)
	}
	return nil
}

func createMultipartFormData(fieldName string, filename string) (io.Reader, string, error) {
	filenameStripped := strings.TrimPrefix(filename, "/tmp/")
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	file, err := os.Open(filename)
	if err != nil {
		return nil, "", fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()
	fw, err := w.CreateFormFile(fieldName, filenameStripped)
	if err != nil {
		return nil, "", fmt.Errorf("error creating writer: %s", err)
	}
	if _, err = io.Copy(fw, file); err != nil {
		return nil, "", fmt.Errorf("error with io.Copy: %s", err)
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &b, w.FormDataContentType(), nil
}

package api

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ApiClient interface {
	Do(request HciRequest) (*HciResponse, error)
	GetApiURL() string
	GetApiKey() string
}

type HciApiClient struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

const API_KEY_HEADER = "MC-Api-Key"

func NewApiClient(apiURL, apiKey string) ApiClient {
	return HciApiClient{
		apiURL:     apiURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func NewInsecureApiClient(apiURL, apiKey string) ApiClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return HciApiClient{
		apiURL:     apiURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Transport: tr},
	}
}

//Build a URL by using endpoint and options. Options will be set as query parameters.
func (hciClient HciApiClient) buildUrl(endpoint string, options map[string]string) string {
	query := url.Values{}
	if options != nil {
		for k, v := range options {
			query.Add(k, v)
		}
	}
	u, _ := url.Parse(hciClient.apiURL + "/" + strings.Trim(endpoint, "/") + "?" + query.Encode())
	return u.String()
}

//Does the API call to server and returns a HCIResponse. hci errors will be returned in the
//HCIResponse body, not in the error return value. The error return value is reserved for unexpected errors.
func (hciClient HciApiClient) Do(request HciRequest) (*HciResponse, error) {
	var bodyBuffer io.Reader
	if request.Body != nil {
		bodyBuffer = bytes.NewBuffer(request.Body)
	}
	method := request.Method
	if method == "" {
		method = "GET"
	}
	req, err := http.NewRequest(request.Method, hciClient.buildUrl(request.Endpoint, request.Options), bodyBuffer)
	if err != nil {
		return nil, err
	}
	req.Header.Add(API_KEY_HEADER, hciClient.apiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := hciClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return NewHciResponse(resp)
}

func (hciClient HciApiClient) GetApiKey() string {
	return hciClient.apiKey
}

func (hciClient HciApiClient) GetApiURL() string {
	return hciClient.apiURL
}

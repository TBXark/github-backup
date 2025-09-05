package request

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

func DefaultHttpClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

func URL(base string, query map[string]string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	params := url.Values{}
	for k, v := range query {
		params.Add(k, v)
	}
	baseURL.RawQuery = params.Encode()
	return baseURL.String(), nil
}

type Modifier func(client *http.Client, req *http.Request)

func Request(method, url string, modifier ...Modifier) (*http.Response, error) {
	client := DefaultHttpClient()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	for _, m := range modifier {
		m(client, req)
	}
	return client.Do(req)
}

func GET[T any](url string, modifier ...Modifier) (*T, error) {
	client := DefaultHttpClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for _, m := range modifier {
		m(client, req)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	var result T
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func POST[T any](url string, data any, modifier ...Modifier) (*T, error) {
	client := DefaultHttpClient()
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	for _, m := range modifier {
		m(client, req)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	var result T
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func WithAuthorization(token string, prefix string) Modifier {
	return func(client *http.Client, req *http.Request) {
		req.Header.Add("Authorization", prefix+" "+token)
	}
}

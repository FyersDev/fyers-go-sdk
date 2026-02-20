package fyersgosdk

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type HTTPClient interface {
	Do(method, rURL string, params url.Values, headers http.Header) (HTTPResponse, error)
	DoRaw(method, rURL string, reqBody []byte, headers http.Header) (HTTPResponse, error)

	GetClient() *httpClient
}

type httpClient struct {
	client *http.Client
	hLog   *log.Logger
	debug  bool
}

type HTTPResponse struct {
	Body     []byte
	Response *http.Response
}

func NewHTTPClient(h *http.Client, hLog *log.Logger, debug bool) HTTPClient {
	if hLog == nil {
		hLog = log.New(os.Stdout, "base.HTTP: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	if h == nil {
		h = &http.Client{
			Timeout: time.Duration(5) * time.Second,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Second * time.Duration(5),
			},
		}
	}

	return &httpClient{
		hLog:   hLog,
		client: h,
		debug:  debug,
	}
}

func (h *httpClient) Do(method, rURL string, params url.Values, headers http.Header) (HTTPResponse, error) {
	if params == nil {
		params = url.Values{}
	}

	return h.DoRaw(method, rURL, []byte(params.Encode()), headers)
}

func (h *httpClient) DoRaw(method, rURL string, reqBody []byte, headers http.Header) (HTTPResponse, error) {
	var (
		resp     = HTTPResponse{}
		err      error
		postBody io.Reader
	)

	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
		postBody = bytes.NewReader(reqBody)
	}

	req, err := http.NewRequest(method, rURL, postBody)
	if err != nil {
		h.hLog.Printf("Request preparation failed: %v", err)
	}

	if headers != nil {
		req.Header = headers
	}

	if req.Header.Get("Content-Type") == "" {
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	if method == http.MethodGet || (method == http.MethodDelete && req.Header.Get("Content-Type") != "application/json") {
		req.URL.RawQuery = string(reqBody)
	}

	r, err := h.client.Do(req)
	if err != nil {
		h.hLog.Printf("Request failed: %v", err)
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.hLog.Printf("Unable to read response: %v", err)
	}

	resp.Response = r
	resp.Body = body
	if h.debug {
		h.hLog.Printf("%s %s -- %d %v", method, req.URL.RequestURI(), resp.Response.StatusCode, req.Header)
	}

	return resp, nil
}

func (h *httpClient) GetClient() *httpClient {
	return h
}

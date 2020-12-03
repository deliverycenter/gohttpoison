package gohttpoison

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

type HTTPClient struct {
	maxLogsBodyChars int
}

func New(maxLogsBodyChars int) *HTTPClient {
	return &HTTPClient{
		maxLogsBodyChars: maxLogsBodyChars,
	}
}

func (h *HTTPClient) Request(r *Request) (*Response, error) {
	var body io.Reader
	var bodyString string

	if r.Body != nil {
		// Marshall payload to json
		requestPayload, err := json.Marshal(r.Body)
		if err != nil {
			return nil, err
		}

		if r.LogRequestBody {
			// prepare payload to print on log
			bodyString = string(requestPayload)
			if len(bodyString) > h.maxLogsBodyChars {
				bodyString = bodyString[0:h.maxLogsBodyChars]
			}
		}

		body = bytes.NewBuffer(requestPayload)
	}

	logger := logrus.WithFields(logrus.Fields{
		"Method": r.Method,
		"URL":    r.URL,
		"Body":   bodyString,
	})
	logger.Debug("HTTPClient request")

	req, err := http.NewRequest(r.Method, r.URL, body)
	if err != nil {
		return nil, err
	}

	// Set URL query params
	q := url.Values(r.Params)
	req.URL.RawQuery = q.Encode()

	// Set headers
	if r.Headers != nil {
		req.Header = r.Headers
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	bodyString = ""
	if r.LogResponseBody {
		bodyString = string(respBytes)
		if len(bodyString) > h.maxLogsBodyChars {
			bodyString = bodyString[0:h.maxLogsBodyChars]
		}
	}
	logrus.WithFields(logrus.Fields{
		"Status-Code": resp.StatusCode,
		"Body":        bodyString,
	}).Debug("HTTPClient response")

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBytes,
		Headers:    resp.Header,
		Request:    r,
	}, nil
}

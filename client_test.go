package gohttpoison

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	maxLogsBodyChars = 10000
)

func TestRequestWithStructBody(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	request := &Request{
		Body: struct {
			Test string `json:"test"`
		}{
			Test: "test",
		},
	}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			assert.JSONEq(t, `{"test":"test"}`, string(body))
			w.WriteHeader(http.StatusOK)
		}),
	)
	request.URL = ts.URL

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Body)
}

func TestRequestWithHugeBody(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	hugeString := "a"
	for i := 0; i < maxLogsBodyChars+1; i++ {
		hugeString += "a"
	}

	request := &Request{
		Body: struct {
			Test string `json:"test"`
		}{hugeString},
		LogRequestBody: true,
	}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			expected := `{"test": "` + hugeString + `"}`
			assert.JSONEq(t, expected, string(body))
			w.WriteHeader(http.StatusOK)
		}),
	)
	request.URL = ts.URL

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Body)
}

func TestRequestWithNilBody(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	request := &Request{}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.NoBody, r.Body)
			w.WriteHeader(http.StatusOK)
		}),
	)
	request.URL = ts.URL

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Body)
}

func TestRequestWithInvalidBody(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	test := make(chan int)
	request := &Request{
		Body: test,
	}

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.EqualError(t, err, "json: unsupported type: chan int")
	assert.Nil(t, resp)
}

func TestRequestWithInvalidMethod(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	request := &Request{
		Method: "(INVALID)",
	}

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.EqualError(t, err, `net/http: invalid method "(INVALID)"`)
	assert.Nil(t, resp)
}

func TestRequestWithoutHeaders(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	request := &Request{}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
		}),
	)
	request.URL = ts.URL

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestWithHeaders(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	request := &Request{
		Headers: map[string][]string{
			"Authorization": {"Bearer token"},
		},
	}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
		}),
	)
	request.URL = ts.URL

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestWithInvalidURL(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	request := &Request{}

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.EqualError(t, err, `Get "": unsupported protocol scheme ""`)
	assert.Nil(t, resp)
}

func TestRequestWhenValidResponse(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	request := &Request{}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"test":"test"}`))
		}),
	)
	request.URL = ts.URL

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Headers["Content-Type"][0])
	assert.JSONEq(t, `{"test":"test"}`, string(resp.Body))
	assert.EqualValues(t, request, resp.Request)
}

func TestRequestWhenHugeResponse(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)

	hugeString := "a"
	for i := 0; i < maxLogsBodyChars+1; i++ {
		hugeString += "a"
	}

	request := &Request{
		LogResponseBody: true,
	}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"test":"` + hugeString + `"}`))
		}),
	)
	request.URL = ts.URL

	httpClient := New(maxLogsBodyChars)
	resp, err := httpClient.Request(request)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"test":"`+hugeString+`"}`, string(resp.Body))
	assert.EqualValues(t, request, resp.Request)
}

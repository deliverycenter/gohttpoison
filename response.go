package gohttpoison

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
	Request    *Request
}

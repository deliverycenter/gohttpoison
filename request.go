package gohttpoison

type Request struct {
	Method          string
	URL             string
	Body            interface{}
	Headers         map[string][]string
	Params          map[string][]string
	LogRequestBody  bool
	LogResponseBody bool
}

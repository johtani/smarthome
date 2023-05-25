package internal

import "net/http"

func BuildHttpRequestWithParams(method string, url string, params map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	for key, param := range params {
		q.Set(key, param)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}

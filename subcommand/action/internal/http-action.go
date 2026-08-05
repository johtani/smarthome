package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MaxErrorResponseBodySize is the maximum number of response body bytes retained for diagnostics.
const MaxErrorResponseBodySize = 4 * 1024

// HTTPStatusError describes an HTTP response with an unexpected status code.
type HTTPStatusError struct {
	StatusCode    int
	ContentType   string
	ResponseBody  string
	BodyTruncated bool
}

// Error returns a concise error message that is safe to propagate to callers.
func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
}

// BuildHttpRequestWithParams creates an HTTP request with the given context, method, URL, and query parameters.
func BuildHttpRequestWithParams(ctx context.Context, method string, url string, params map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
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

// HandleResponse checks if the response status code is among the expected statuses and closes the response body.
func HandleResponse(res *http.Response, expectedStatuses ...int) error {
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()

	isExpected := false
	if len(expectedStatuses) == 0 {
		// デフォルトで 200 OK を期待する場合
		if res.StatusCode == http.StatusOK {
			isExpected = true
		}
	} else {
		for _, status := range expectedStatuses {
			if res.StatusCode == status {
				isExpected = true
				break
			}
		}
	}

	if !isExpected {
		body, truncated, err := readErrorResponseBody(res.Body)
		if err != nil {
			return fmt.Errorf("read unexpected HTTP response body: %w", err)
		}
		return &HTTPStatusError{
			StatusCode:    res.StatusCode,
			ContentType:   res.Header.Get("Content-Type"),
			ResponseBody:  body,
			BodyTruncated: truncated,
		}
	}
	return nil
}

func readErrorResponseBody(body io.Reader) (string, bool, error) {
	limited := io.LimitReader(body, MaxErrorResponseBodySize+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return "", false, err
	}
	truncated := len(b) > MaxErrorResponseBodySize
	if truncated {
		b = b[:MaxErrorResponseBodySize]
	}
	return strings.TrimSpace(strings.ToValidUTF8(string(b), "�")), truncated, nil
}

// DecodeJSONResponse checks the response status code and decodes the JSON response body into the target.
// It also ensures the response body is closed.
func DecodeJSONResponse[T any](res *http.Response, target *T, expectedStatuses ...int) error {
	// HandleResponse と同様の defer 処理が必要だが、
	// HandleResponse を呼ぶと Body が閉じられてしまうので、ここではインラインで書くか工夫が必要。
	// ここでは、ステータスチェック後にデコードし、最後に閉じるようにする。

	isExpected := false
	if len(expectedStatuses) == 0 {
		if res.StatusCode == http.StatusOK {
			isExpected = true
		}
	} else {
		for _, status := range expectedStatuses {
			if res.StatusCode == status {
				isExpected = true
				break
			}
		}
	}

	if !isExpected {
		defer func() {
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()
		}()
		return fmt.Errorf("unexpected status code: %d, header: %v", res.StatusCode, res.Header)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()

	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

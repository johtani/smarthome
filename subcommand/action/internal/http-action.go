package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
		return fmt.Errorf("unexpected status code: %d, header: %v", res.StatusCode, res.Header)
	}
	return nil
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

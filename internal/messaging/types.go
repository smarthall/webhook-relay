package messaging

import (
	"bytes"
	"io"
	"net/http"
)

type RequestMessage struct {
	Method  string              `json:"method"`
	Host    string              `json:"host"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

// FromHTTPRequest populates the RequestMessage from an http.Request.
// It reads the request body (consuming it) and copies method, host, path and headers.
func (rm *RequestMessage) FromHTTPRequest(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	// It's safe to close here; caller may not expect Body to be usable after.
	if rc, ok := interface{}(r.Body).(io.ReadCloser); ok {
		rc.Close()
	}

	rm.Method = r.Method
	rm.Host = r.Host
	if r.URL != nil {
		rm.Path = r.URL.Path
	}
	rm.Headers = r.Header
	rm.Body = string(body)
	return nil
}

// ToHTTPRequest builds an *http.Request from the RequestMessage targeting destURL.
// The returned request will have headers and body set. Caller should handle
// additional relay headers or Host preservation as desired.
func (rm *RequestMessage) ToHTTPRequest(destURL string) (*http.Request, error) {
	buf := bytes.NewBufferString(rm.Body)
	req, err := http.NewRequest(rm.Method, destURL, buf)
	if err != nil {
		return nil, err
	}

	// copy headers
	if rm.Headers != nil {
		for k, v := range rm.Headers {
			req.Header[k] = append([]string(nil), v...)
		}
	}

	return req, nil
}

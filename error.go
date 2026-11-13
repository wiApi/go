package wi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error is returned when the API responds with a 4xx or 5xx status.
type Error struct {
	Status  int
	Message string
	Code    string
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("wi: HTTP %d %s (%s)", e.Status, e.Message, e.Code)
	}
	return fmt.Sprintf("wi: HTTP %d %s", e.Status, e.Message)
}

func parseErrorResponse(res *http.Response) *Error {
	e := &Error{Status: res.StatusCode, Message: http.StatusText(res.StatusCode)}

	// Current shape: {"error":{"code","message"}}. The legacy flat shape
	// {"error":"...","message":"...","code":"..."} is still read.
	var body struct {
		Message string          `json:"message"`
		Error   json.RawMessage `json:"error"`
		Code    string          `json:"code"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err == nil {
		var nested struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		var flat string
		switch {
		case json.Unmarshal(body.Error, &nested) == nil && (nested.Code != "" || nested.Message != ""):
			if nested.Message != "" {
				e.Message = nested.Message
			}
			e.Code = nested.Code
		default:
			if body.Message != "" {
				e.Message = body.Message
			} else if json.Unmarshal(body.Error, &flat) == nil && flat != "" {
				e.Message = flat
			}
			e.Code = body.Code
		}
	}
	return e
}

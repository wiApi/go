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

	var body struct {
		Message string `json:"message"`
		Error   string `json:"error"`
		Code    string `json:"code"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err == nil {
		if body.Message != "" {
			e.Message = body.Message
		} else if body.Error != "" {
			e.Message = body.Error
		}
		e.Code = body.Code
	}
	return e
}

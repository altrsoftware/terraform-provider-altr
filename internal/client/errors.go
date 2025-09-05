package client

import (
	"fmt"
)

type APIError struct {
	StatusCode int
	Response   APIErrorResponse `json:"error"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("Code %d: %s", e.Response.ErrorCode, e.Response.Message)
}

type APIErrorResponse struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}

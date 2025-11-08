package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sog01/diycctv/pkg/errors"
	"github.com/sog01/diycctv/pkg/logger"

	appctx "github.com/sog01/diycctv/pkg/context"
)

// Response represents a standard API response structure with request ID
type Response struct {
	Data      any    `json:"data,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorResponse represents a standard error response for the API
// swagger:response ErrorResponse
type ErrorResponse struct {
	// The request ID for tracing
	// in: body
	RequestID string `json:"request_id,omitempty"`

	// Error message
	// in: body
	Error string `json:"error,omitempty"`

	// HTTP status code
	// in: body
	Code int `json:"code,omitempty"`
}

func (r *Response) WriteSuccessJSON(w http.ResponseWriter, ctx context.Context, statusCode int, data any) {
	writeJSON(w, ctx, statusCode, data, nil)
}

func (r *Response) WriteErrorJSON(w http.ResponseWriter, ctx context.Context, statusCode int, err error) {
	writeJSON(w, ctx, statusCode, nil, err)
}

func (r *Response) WriteWithAppErrorJSON(w http.ResponseWriter, ctx context.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		writeJSON(w, ctx, appErr.StatusCode, nil, err)
	}
}

func writeJSON(w http.ResponseWriter, ctx context.Context, statusCode int, data any, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Extract request ID from context
	requestID := appctx.GetRequestID(ctx)

	resp := Response{}
	if data != nil {
		resp.Data = data
	}
	// Always include the request ID in responses
	resp.RequestID = requestID
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Message = "success"
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.ErrorLogger.Err(err).Msgf("error encoding JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

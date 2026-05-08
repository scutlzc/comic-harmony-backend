package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, APIResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, APIResponse{
		Code:    201,
		Message: "created",
		Data:    data,
	})
}

func NotFound(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "not found"
	}
	JSON(w, http.StatusNotFound, APIResponse{
		Code:    404,
		Message: msg,
	})
}

func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, APIResponse{
		Code:    400,
		Message: msg,
	})
}

func Unauthorized(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "unauthorized"
	}
	JSON(w, http.StatusUnauthorized, APIResponse{
		Code:    401,
		Message: msg,
	})
}

func InternalError(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "internal server error"
	}
	JSON(w, http.StatusInternalServerError, APIResponse{
		Code:    500,
		Message: msg,
	})
}

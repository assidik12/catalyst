package response

import (
	"encoding/json"
	"net/http"
)

// WebResponse adalah struktur standar untuk semua respons JSON dari API.
type WebResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

// writeResponse adalah fungsi internal untuk menghindari duplikasi kode.
func writeResponse(w http.ResponseWriter, code int, status string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(WebResponse{
		Code:   code,
		Status: status,
		Data:   data,
	})
}

// OK mengirim respons 200 OK dengan data.
func OK(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusOK, "OK", data)
}

// Created mengirim respons 201 Created dengan data.
func Created(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusCreated, "CREATED", data)
}

// BadRequest mengirim respons 400 Bad Request dengan pesan error.
func BadRequest(w http.ResponseWriter, errorMessage string) {
	writeResponse(w, http.StatusBadRequest, "BAD REQUEST", map[string]string{"error": errorMessage})
}

// NotFound mengirim respons 404 Not Found dengan pesan error.
func NotFound(w http.ResponseWriter, errorMessage string) {
	writeResponse(w, http.StatusNotFound, "NOT FOUND", map[string]string{"error": errorMessage})
}

// InternalServerError mengirim respons 500 Internal Server Error dengan pesan error.
func InternalServerError(w http.ResponseWriter, errorMessage string) {
	writeResponse(w, http.StatusInternalServerError, "INTERNAL SERVER ERROR", map[string]string{"error": errorMessage})
}

func Unauthorized(w http.ResponseWriter, errorMessage string) {
	writeResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", map[string]string{"error": errorMessage})
}

// Conflict sends a 409 Conflict response (e.g. duplicate resource).
func Conflict(w http.ResponseWriter, errorMessage string) {
	writeResponse(w, http.StatusConflict, "CONFLICT", map[string]string{"error": errorMessage})
}

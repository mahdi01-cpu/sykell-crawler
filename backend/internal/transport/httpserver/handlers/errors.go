package handlers

import (
	"errors"
	"net/http"

	"github.com/mahdi-01/sykell-crawler/internal/domain"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidURL),
		errors.Is(err, domain.ErrInvalidURLStatus):
		writeJSON(w, http.StatusBadRequest, apiError{Code: "bad_request", Message: err.Error()})

	case errors.Is(err, domain.ErrAlreadyExists):
		writeJSON(w, http.StatusConflict, apiError{Code: "already_exists", Message: err.Error()})

	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, apiError{Code: "not_found", Message: err.Error()})

	default:
		writeJSON(w, http.StatusInternalServerError, apiError{Code: "internal", Message: "internal server error"})
	}
}

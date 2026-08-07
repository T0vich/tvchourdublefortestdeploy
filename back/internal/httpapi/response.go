package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"trade-chain/internal/service"
)

type ErrorResponse struct {
	Error string `json:"error"`
	// Fields заполняется только для ошибок валидации: "поле -> причина".
	Fields map[string]string `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrNotFound):
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrConflict):
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrForbidden):
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: err.Error()})
	default:
		// Незнакомая ошибка — это почти всегда ошибка базы или драйвера.
		// Её текст содержит имена таблиц, колонок и фрагменты запроса,
		// поэтому наружу уходит нейтральное сообщение, а подробности в лог.
		log.Printf("необработанная ошибка: %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "внутренняя ошибка сервера"})
	}
}

// writeValidationError отдаёт 422 с разбором по полям.
func writeValidationError(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
		Error:  "переданные данные не прошли проверку",
		Fields: fields,
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

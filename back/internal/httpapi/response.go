package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"unicode/utf8"

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
	case errors.Is(err, service.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "неверный email или пароль"})
	case errors.Is(err, service.ErrForbidden):
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "недостаточно прав для этого действия"})
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

// maxBodySize ограничивает размер тела запроса. Без ограничения любой клиент
// может занять память сервера одним запросом.
const maxBodySize = 1 << 20 // 1 МБ

func decodeJSON(r *http.Request, dst any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		return err
	}

	// encoding/json молча заменяет некорректные байты на U+FFFD, поэтому текст
	// в другой кодировке сохраняется как строка ромбиков вместо ошибки.
	// Проверяем до разбора, пока байты ещё исходные.
	if !utf8.Valid(body) {
		return errInvalidEncoding
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()

	return dec.Decode(dst)
}

var errInvalidEncoding = errors.New("тело запроса не является корректным UTF-8")

// decodeBody разбирает тело запроса и сам отвечает при неудаче.
// Возвращает false, если обработчику следует прекратить работу.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	err := decodeJSON(r, dst)
	switch {
	case err == nil:
		return true
	case errors.Is(err, errInvalidEncoding):
		writeValidationError(w, map[string]string{"body": err.Error()})
	default:
		writeError(w, service.ErrInvalidInput)
	}

	return false
}

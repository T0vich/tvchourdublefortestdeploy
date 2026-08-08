package httpapi

import (
	"net/http"
	"net/mail"
	"unicode/utf8"
)

// requireUTF8Query отклоняет запросы, в которых параметры строки запроса
// после декодирования не являются корректным UTF-8.
//
// Без этой проверки такие байты доезжают до Postgres и он отвечает
// "invalid byte sequence for encoding UTF8" — то есть ошибка клиента
// превращается в 500. Ловится это, например, когда консоль отправляет
// кириллицу в CP1251.
func requireUTF8Query(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, values := range r.URL.Query() {
			if !utf8.ValidString(key) {
				writeValidationError(w, map[string]string{"query": "имя параметра не является корректным UTF-8"})
				return
			}
			for _, value := range values {
				if !utf8.ValidString(value) {
					writeValidationError(w, map[string]string{key: "значение не является корректным UTF-8"})
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// minPasswordLength повторяет ограничение из тегов validate в domain:
// теги там декоративные — в проекте нет библиотеки, которая их читает,
// поэтому проверка живёт на транспортном слое.
const minPasswordLength = 8

// validateCredentials возвращает карту "поле -> причина отказа".
// Пустая карта означает, что данные приняты.
func validateCredentials(email, password string) map[string]string {
	fields := make(map[string]string)

	switch {
	case email == "":
		fields["email"] = "обязательное поле"
	default:
		if _, err := mail.ParseAddress(email); err != nil {
			fields["email"] = "некорректный адрес электронной почты"
		}
	}

	switch {
	case password == "":
		fields["password"] = "обязательное поле"
	case utf8.RuneCountInString(password) < minPasswordLength:
		fields["password"] = "минимальная длина — 8 символов"
	}

	return fields
}

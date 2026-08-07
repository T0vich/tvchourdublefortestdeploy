package httpapi

import (
	"net/mail"
	"unicode/utf8"
)

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

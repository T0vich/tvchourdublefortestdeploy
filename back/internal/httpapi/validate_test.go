package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		password   string
		wantFields []string
	}{
		{"корректные данные", "anna@example.com", "password123", nil},
		{"пустой email", "", "password123", []string{"email"}},
		{"email без домена", "anna", "password123", []string{"email"}},
		{"пустой пароль", "anna@example.com", "", []string{"password"}},
		{"короткий пароль", "anna@example.com", "1234567", []string{"password"}},
		{"пароль ровно 8", "anna@example.com", "12345678", nil},
		{"кириллица считается по символам, а не байтам", "anna@example.com", "парольпароль", nil},
		{"обе ошибки сразу", "нет", "1", []string{"email", "password"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateCredentials(tt.email, tt.password)

			if len(got) != len(tt.wantFields) {
				t.Fatalf("получено %v, ожидались поля %v", got, tt.wantFields)
			}
			for _, field := range tt.wantFields {
				if _, ok := got[field]; !ok {
					t.Errorf("ожидалась ошибка по полю %q, получено %v", field, got)
				}
			}
		})
	}
}

func TestRequireUTF8Query(t *testing.T) {
	handler := requireUTF8Query(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("корректный UTF-8 проходит", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		// %D0%BD%D0%BE%D1%83%D1%82 — "ноут" в UTF-8
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/products/search?q=%D0%BD%D0%BE%D1%83%D1%82", nil)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("получен код %d, ожидался 200", recorder.Code)
		}
	})

	t.Run("байты CP1251 отклоняются с 422, а не доезжают до базы", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		// %ED%EE%F3 — те же "ноу", но в CP1251: некорректная последовательность UTF-8
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/products/search?q=%ED%EE%F3", nil)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnprocessableEntity {
			t.Fatalf("получен код %d, ожидался 422", recorder.Code)
		}

		var body ErrorResponse
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("не удалось разобрать тело ответа: %s", err)
		}
		if _, ok := body.Fields["q"]; !ok {
			t.Errorf("ожидалась ошибка по полю q, получено %v", body.Fields)
		}
	})
}

package httpapi

import (
	"net/http"
	"os"
	"slices"
	"strings"
)

const requestIDHeader = "X-Request-Id"

// allowedOrigins читает список источников из CORS_ALLOWED_ORIGINS
// (значения через запятую). Пустое значение означает "*": фронт и бэк
// в этом проекте живут на разных доменах, и без разрешённого источника
// браузер отклонит вообще каждый запрос.
func allowedOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		return []string{"*"}
	}

	var origins []string
	for _, origin := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return origins
}

// cors проставляет заголовки Access-Control-* и отвечает на preflight-запросы.
//
// Источник всегда отражается в ответ поимённо, даже когда разрешены любые:
// со звёздочкой в Access-Control-Allow-Origin браузер запрещает
// Access-Control-Allow-Credentials, а он нужен для передачи токена.
func cors(origins []string) func(http.Handler) http.Handler {
	allowAny := slices.Contains(origins, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			w.Header().Add("Vary", "Origin")

			if origin != "" && (allowAny || slices.Contains(origins, origin)) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{
					http.MethodGet,
					http.MethodPost,
					http.MethodPatch,
					http.MethodPut,
					http.MethodDelete,
					http.MethodOptions,
				}, ", "))
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, "+requestIDHeader)
				w.Header().Set("Access-Control-Expose-Headers", requestIDHeader)
				w.Header().Set("Access-Control-Max-Age", "600")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

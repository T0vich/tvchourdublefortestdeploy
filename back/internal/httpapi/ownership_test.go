package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/service"
)

// stubProducts реализует ровно тот кусок service.ProductService, который нужен
// проверкам владельца; остальные методы не вызываются.
type stubProducts struct {
	service.ProductService
	byID map[string]*domain.Product
}

func (s stubProducts) GetByID(_ context.Context, id string) (*domain.Product, error) {
	product, ok := s.byID[id]
	if !ok {
		return nil, service.ErrNotFound
	}

	return product, nil
}

// withActor кладёт идентификатор пользователя в контекст так же, как это делает
// auth.AuthMiddleware после разбора токена.
func withActor(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), auth.UserIDKey, userID))
}

func TestRequireSelf(t *testing.T) {
	tests := []struct {
		name     string
		actorID  string
		targetID string
		wantCode int
		wantOK   bool
	}{
		{"свой аккаунт", "user-1", "user-1", http.StatusOK, true},
		{"чужой аккаунт", "user-1", "user-2", http.StatusForbidden, false},
		{"без токена", "", "user-2", http.StatusUnauthorized, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/customers/"+tt.targetID, nil)
			if tt.actorID != "" {
				request = withActor(request, tt.actorID)
			}

			ok := requireSelf(recorder, request, tt.targetID)

			if ok != tt.wantOK {
				t.Fatalf("получено ok=%v, ожидалось %v", ok, tt.wantOK)
			}
			if !tt.wantOK && recorder.Code != tt.wantCode {
				t.Errorf("получен код %d, ожидался %d", recorder.Code, tt.wantCode)
			}
		})
	}
}

func TestRequireProductOwner(t *testing.T) {
	products := stubProducts{byID: map[string]*domain.Product{
		"product-1": {ProductID: "product-1", CustomerID: "user-1", Title: "iPhone 15"},
	}}

	tests := []struct {
		name      string
		actorID   string
		productID string
		wantCode  int
		wantOK    bool
	}{
		{"владелец правит свой товар", "user-1", "product-1", http.StatusOK, true},
		{"чужой товар не отдаётся", "user-2", "product-1", http.StatusForbidden, false},
		{"несуществующий товар", "user-1", "product-404", http.StatusNotFound, false},
		{"без токена", "", "product-1", http.StatusUnauthorized, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/products/"+tt.productID, nil)
			if tt.actorID != "" {
				request = withActor(request, tt.actorID)
			}

			ok := requireProductOwner(recorder, request, products, tt.productID)

			if ok != tt.wantOK {
				t.Fatalf("получено ok=%v, ожидалось %v", ok, tt.wantOK)
			}
			if !tt.wantOK && recorder.Code != tt.wantCode {
				t.Errorf("получен код %d, ожидался %d", recorder.Code, tt.wantCode)
			}
		})
	}
}

// Ответ при отказе не должен раскрывать, существует ли объект и кому он
// принадлежит, — наружу уходит только общая формулировка.
func TestForbiddenResponseDoesNotLeakDetails(t *testing.T) {
	products := stubProducts{byID: map[string]*domain.Product{
		"product-1": {ProductID: "product-1", CustomerID: "user-1", Title: "iPhone 15"},
	}}

	recorder := httptest.NewRecorder()
	request := withActor(httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/products/product-1", nil), "user-2")

	requireProductOwner(recorder, request, products, "product-1")

	body := recorder.Body.String()
	for _, secret := range []string{"user-1", "iPhone 15"} {
		if strings.Contains(body, secret) {
			t.Errorf("в ответе просочилось %q: %s", secret, body)
		}
	}
}

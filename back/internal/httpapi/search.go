package httpapi

import (
	"net/http"
	"strconv"
	"trade-chain/internal/auth"
	"trade-chain/internal/search"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

type searchHandler struct {
	s *search.SearchService
}

func mountSearchRoutes(r chi.Router, s *search.SearchService) {
	h := searchHandler{s}
	// Поиск цепочки строится от товаров текущего пользователя,
	// поэтому без токена он бессмысленен.
	r.Route("/search", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		r.Get("/chain", h.findChain)
	})
}

// findChain godoc
// @Summary Find exchange chain
// @Description Find a chain of exchanges from user's products to target product
// @Tags search
// @Accept json
// @Produce json
// @Param target_product_id query string true "Target product ID"
// @Param max_depth query int false "Maximum depth" default(10)
// @Success 200 {object} map[string]interface{} "chain (array of products) and length"
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /search/chain [get]
func (h searchHandler) findChain(w http.ResponseWriter, r *http.Request) {
	userID, ok := actor(w, r)
	if !ok {
		return
	}

	targetProductID := r.URL.Query().Get("target_product_id")
	if targetProductID == "" {
		writeError(w, service.ErrInvalidInput)
		return
	}

	maxDepthStr := r.URL.Query().Get("max_depth")
	maxDepth := 10
	if maxDepthStr != "" {
		if d, err := strconv.Atoi(maxDepthStr); err == nil && d > 0 {
			maxDepth = d
		}
	}

	result, err := h.s.FindChain(r.Context(), userID, targetProductID, maxDepth)
	if err != nil {
		writeError(w, err)
		return
	}
	if result == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"chain": []interface{}{}, "length": 0})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"chain":  result.Products,
		"length": result.Length,
	})
}

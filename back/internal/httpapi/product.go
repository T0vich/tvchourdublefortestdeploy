package httpapi

import (
	"net/http"
	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

type productHandler struct{ s service.ProductService }

func mountProductRoutes(r chi.Router, s service.ProductService) {
	h := productHandler{s}
	r.Route("/products", func(r chi.Router) {
		// Каталог и карточки открыты без токена: это витрина.
		r.Get("/", h.list)
		r.Get("/search", h.search)
		r.Get("/{id}", h.get)
		r.Get("/by-customer/{customerID}", h.byCustomer)

		protected := r.With(auth.AuthMiddleware)
		protected.Post("/", h.create)
		protected.Patch("/{id}", h.update)
		protected.Delete("/{id}", h.delete)
	})
}

// create godoc
// @Summary Create product
// @Description Create a new product listing
// @Tags products
// @Accept json
// @Produce json
// @Param request body domain.CreateProductDTO true "Product data"
// @Success 201 {object} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products [post]
func (h productHandler) create(w http.ResponseWriter, r *http.Request) {
	var v domain.CreateProductDTO
	if decodeJSON(r, &v) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	out, e := h.s.Create(r.Context(), &v)
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// get godoc
// @Summary Get product by ID
// @Description Get product details
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [get]
func (h productHandler) get(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetByID(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// update godoc
// @Summary Update product
// @Description Update product information
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param request body domain.UpdateProductDTO true "Updated product data"
// @Success 200 {object} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [patch]
func (h productHandler) update(w http.ResponseWriter, r *http.Request) {
	var v domain.UpdateProductDTO
	if decodeJSON(r, &v) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}
	out, e := h.s.Update(r.Context(), chi.URLParam(r, "id"), &v)
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// delete godoc
// @Summary Delete product
// @Description Soft delete product (set status to archived)
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/{id} [delete]
func (h productHandler) delete(w http.ResponseWriter, r *http.Request) {
	if e := h.s.Delete(r.Context(), chi.URLParam(r, "id")); e != nil {
		writeError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// list godoc
// @Summary List products
// @Description List products with pagination
// @Tags products
// @Accept json
// @Produce json
// @Param offset query int false "Offset" default(0)
// @Param limit query int false "Limit" default(20) maximum(100)
// @Success 200 {array} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products [get]
func (h productHandler) list(w http.ResponseWriter, r *http.Request) {
	o, l, e := pagination(r)
	if e != nil {
		writeError(w, e)
		return
	}
	v, e := h.s.List(r.Context(), o, l)
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// search godoc
// @Summary Search products
// @Description Search products by text query and optionally filter by category
// @Tags products
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param category_id query string false "Category ID"
// @Success 200 {array} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/search [get]
func (h productHandler) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	var category *string
	if v := r.URL.Query().Get("category_id"); v != "" {
		category = &v
	}
	out, e := h.s.Search(r.Context(), q, category)
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// byCustomer godoc
// @Summary Get products by customer
// @Description Get all products owned by a customer
// @Tags products
// @Accept json
// @Produce json
// @Param customerID path string true "Customer ID"
// @Success 200 {array} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /products/by-customer/{customerID} [get]
func (h productHandler) byCustomer(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetByCustomerID(r.Context(), chi.URLParam(r, "customerID"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

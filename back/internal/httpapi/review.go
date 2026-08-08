package httpapi

import (
	"net/http"
	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
)

type reviewHandler struct{ s service.ReviewService }

func mountReviewRoutes(r chi.Router, s service.ReviewService) {
	h := reviewHandler{s}
	r.Route("/reviews", func(r chi.Router) {
		// Отзывы и рейтинг — часть публичного профиля продавца.
		r.Get("/{id}", h.get)
		r.Get("/by-customer/{customerID}", h.byCustomer)
		r.Get("/by-customer/{customerID}/rating", h.rating)

		protected := r.With(auth.AuthMiddleware)
		protected.Post("/", h.create)
		protected.Delete("/{id}", h.delete)
	})
}

// create godoc
// @Summary Create review
// @Description Create a new review for a customer
// @Tags reviews
// @Accept json
// @Produce json
// @Param request body domain.Review true "Review data"
// @Success 201 {object} domain.Review
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews [post]
func (h reviewHandler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := actor(w, r)
	if !ok {
		return
	}

	var v domain.Review
	if decodeJSON(r, &v) != nil {
		writeError(w, service.ErrInvalidInput)
		return
	}

	// Автор отзыва — владелец токена. Иначе отзыв можно оставить от чужого
	// имени, подставив нужный from_customer_id в тело запроса.
	v.FromCustomerID = userID
	if v.ToCustomerID == userID {
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
// @Summary Get review by ID
// @Description Get review details
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} domain.Review
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/{id} [get]
func (h reviewHandler) get(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetByID(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// delete godoc
// @Summary Delete review
// @Description Delete a review
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Success 204 "No content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/{id} [delete]
func (h reviewHandler) delete(w http.ResponseWriter, r *http.Request) {
	if ok := requireReviewAuthor(w, r, h.s, chi.URLParam(r, "id")); !ok {
		return
	}

	if e := h.s.Delete(r.Context(), chi.URLParam(r, "id")); e != nil {
		writeError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// byCustomer godoc
// @Summary Get reviews for a customer
// @Description Get all reviews received by a customer
// @Tags reviews
// @Accept json
// @Produce json
// @Param customerID path string true "Customer ID"
// @Success 200 {array} domain.Review
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/by-customer/{customerID} [get]
func (h reviewHandler) byCustomer(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetByCustomerID(r.Context(), chi.URLParam(r, "customerID"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// rating godoc
// @Summary Get average rating for a customer
// @Description Get average rating of a customer
// @Tags reviews
// @Accept json
// @Produce json
// @Param customerID path string true "Customer ID"
// @Success 200 {object} map[string]float64 "average_rating"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reviews/by-customer/{customerID}/rating [get]
func (h reviewHandler) rating(w http.ResponseWriter, r *http.Request) {
	v, e := h.s.GetAverageRating(r.Context(), chi.URLParam(r, "customerID"))
	if e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"average_rating": v})
}

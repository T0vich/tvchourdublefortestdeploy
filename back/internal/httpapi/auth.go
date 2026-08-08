package httpapi

import (
	"net/http"
	"strings"
	"trade-chain/internal/auth"
	"trade-chain/internal/domain"
	"trade-chain/internal/service"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type authHandler struct {
	customerService service.CustomerService
}

func mountAuthRoutes(r chi.Router, cs service.CustomerService) {
	h := authHandler{cs}
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)

		// Профиль текущего пользователя: заодно позволяет фронту проверить,
		// жив ли сохранённый токен.
		r.With(auth.AuthMiddleware).Get("/me", h.me)
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse — тело ответа /auth/register и /auth/login.
// Форма согласована с веткой register-me-router-logic (Максим), чтобы фронт
// не пришлось переписывать при слиянии.
type AuthResponse struct {
	User  domain.Customer `json:"user"`
	Token string          `json:"token"`
}

// register godoc
// @Summary Register user
// @Description Создаёт пользователя и сразу возвращает JWT — отдельный вход после регистрации не нужен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.CreateCustomerDTO true "Email и пароль (пароль от 8 символов)"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 409 {object} ErrorResponse "Email уже занят"
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h authHandler) register(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCustomerDTO
	if !decodeBody(w, r, &req) {
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if fields := validateCredentials(req.Email, req.Password); len(fields) > 0 {
		writeValidationError(w, fields)
		return
	}

	customer, err := h.customerService.Create(r.Context(), &req)
	if err != nil {
		writeError(w, err)
		return
	}

	token, err := auth.GenerateToken(customer.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{User: *customer, Token: token})
}

// login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Некорректное тело запроса"
// @Failure 401 {object} ErrorResponse "Неверный email или пароль"
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !decodeBody(w, r, &req) {
		return
	}
	// Несуществующий email и неверный пароль отдают одинаковый ответ:
	// иначе по коду ответа можно перебором узнать, кто зарегистрирован.
	customer, err := h.customerService.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, service.ErrUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, service.ErrUnauthorized)
		return
	}
	token, err := auth.GenerateToken(customer.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, AuthResponse{User: *customer, Token: token})
}

// me godoc
// @Summary Current user
// @Description Профиль владельца токена
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Customer
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/me [get]
func (h authHandler) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := actor(w, r)
	if !ok {
		return
	}

	customer, err := h.customerService.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, customer)
}

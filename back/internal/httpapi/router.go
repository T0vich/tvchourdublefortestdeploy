package httpapi

import (
	"net/http"
	"time"
	"trade-chain/internal/search"
	"trade-chain/internal/service"

	_ "trade-chain/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Dependencies struct {
	Customers  service.CustomerService
	Products   service.ProductService
	Chains     service.ChainService
	Reviews    service.ReviewService
	Categories service.CategoryService
	Wishlists  service.WishlistService
	Search     *search.SearchService
}

func NewRouter(d Dependencies) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	// middleware.RealIP намеренно не используется: он подменяет r.RemoteAddr
	// значением из X-Forwarded-For, которое клиент задаёт сам (GHSA-3fxj-6jh8-hvhx).
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(cors(allowedOrigins()))
	r.Use(requireUTF8Query)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Требование авторизации навешивается не на всю группу, а на каждый
	// маршрут отдельно (см. r.With(auth.AuthMiddleware) в mount-функциях):
	// каталог, карточки товаров и публичные профили должны открываться
	// без токена, иначе неавторизованный гость видит только 401.
	r.Route("/api/v1", func(r chi.Router) {
		mountAuthRoutes(r, d.Customers)

		if d.Customers != nil {
			mountCustomerRoutes(r, d.Customers)
		}
		if d.Products != nil {
			mountProductRoutes(r, d.Products)
		}
		if d.Chains != nil {
			mountChainRoutes(r, d.Chains)
		}
		if d.Reviews != nil {
			mountReviewRoutes(r, d.Reviews)
		}
		if d.Categories != nil {
			mountCategoryRoutes(r, d.Categories)
		}
		if d.Wishlists != nil && d.Products != nil {
			mountWishlistRoutes(r, d.Wishlists, d.Products)
		}
		if d.Search != nil {
			mountSearchRoutes(r, d.Search)
		}
	})

	return r
}

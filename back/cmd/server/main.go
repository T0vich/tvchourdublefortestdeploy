package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"trade-chain/docs"
	"trade-chain/internal/httpapi"
	"trade-chain/internal/repository"
	"trade-chain/internal/search"
	"trade-chain/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

// @title Trade Chain API
// @version 1.0
// @description API для обмена товарами
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

const (
	defaultPort     = "8080"
	defaultMaxConns = 5

	readHeaderTimeout = 10 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 15 * time.Second
)

func main() {
	ctx := context.Background()

	pool, err := connectDB(ctx)
	if err != nil {
		log.Fatalf("база данных недоступна: %s", err)
	}
	defer pool.Close()

	customerRepo := repository.NewCustomerRepository(pool)
	productRepo := repository.NewProductRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	wishlistRepo := repository.NewWishlistRepository(pool)
	chainRepo := repository.NewChainRepository(pool)
	reviewRepo := repository.NewReviewRepository(pool)

	customerService := service.NewCustomerService(customerRepo)
	productService := service.NewProductService(productRepo, customerRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	wishlistService := service.NewWishlistService(wishlistRepo, productRepo)
	chainService := service.NewChainService(chainRepo, productRepo)
	reviewService := service.NewReviewService(reviewRepo, customerRepo, productRepo)
	searchService := search.NewSearchService(productService, categoryService)

	configureSwagger()

	router := httpapi.NewRouter(httpapi.Dependencies{
		Customers:  customerService,
		Products:   productService,
		Chains:     chainService,
		Reviews:    reviewService,
		Categories: categoryService,
		Wishlists:  wishlistService,
		Search:     searchService,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Таймауты обязательны: http.ListenAndServe без них оставляет соединение
	// открытым бесконечно, и одного медленного клиента хватает, чтобы занять
	// сокет навсегда.
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("сервер слушает порт %s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("сервер остановлен с ошибкой: %s", err)
	case <-stop:
		log.Print("получен сигнал остановки, доигрываем текущие запросы")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("не удалось остановиться мягко: %s", err)
	}
}

// connectDB поднимает пул с ограничением на число соединений: в serverless-среде
// параллельных экземпляров может быть много, а лимит соединений у базы один
// на всех, и пул по умолчанию его быстро выбирает.
func connectDB(ctx context.Context) (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL не задан")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	config.MaxConns = defaultMaxConns
	if raw := os.Getenv("DB_MAX_CONNS"); raw != "" {
		parsed, convErr := strconv.ParseInt(raw, 10, 32)
		if convErr != nil || parsed <= 0 {
			return nil, errors.New("DB_MAX_CONNS должен быть положительным числом")
		}
		config.MaxConns = int32(parsed)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

// configureSwagger подставляет реальный адрес развёрнутого сервиса: в
// сгенерированной спецификации зашит localhost:8080, и на деплое кнопка
// «Try it out» без этого бьёт в машину того, кто открыл страницу.
func configureSwagger() {
	docs.SwaggerInfo.BasePath = "/api/v1"

	host := os.Getenv("PUBLIC_HOST")
	if host == "" {
		host = os.Getenv("VERCEL_URL") // Vercel отдаёт домен без схемы
	}

	if host == "" {
		docs.SwaggerInfo.Host = "localhost:" + cmpOr(os.Getenv("PORT"), defaultPort)
		docs.SwaggerInfo.Schemes = []string{"http"}
		return
	}

	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.Schemes = []string{"https"}
}

func cmpOr(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

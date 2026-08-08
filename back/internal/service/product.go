package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"trade-chain/internal/domain"
	"trade-chain/internal/repository"
)

type productService struct {
	repo      repository.ProductRepository
	customers repository.CustomerRepository
}

func NewProductService(repo repository.ProductRepository, customers repository.CustomerRepository) ProductService {
	return &productService{repo: repo, customers: customers}
}

// Create – создание нового продукта с валидацией и нормализацией ошибок
func (s *productService) Create(ctx context.Context, dto *domain.CreateProductDTO) (*domain.Product, error) {
	if dto == nil || blank(dto.CustomerID) || blank(dto.Title) {
		return nil, ErrInvalidInput
	}
	// Проверяем существование покупателя
	if _, err := s.customers.GetByID(ctx, dto.CustomerID); err != nil {
		return nil, normalizeError(err)
	}

	// Очистка и подготовка данных
	clean := *dto
	clean.Title = strings.TrimSpace(clean.Title)
	clean.Description = strings.TrimSpace(clean.Description)
	clean.Image = strings.TrimSpace(clean.Image)
	clean.Location = strings.TrimSpace(clean.Location)

	// Статус по умолчанию – active, если не передан
	if clean.Status == nil {
		defaultStatus := domain.ProductActive
		clean.Status = &defaultStatus
	}
	if !isValidProductStatus(*clean.Status) {
		return nil, ErrInvalidInput
	}
	// Запрещаем создание продукта со статусом archived
	if *clean.Status == domain.ProductArchived {
		return nil, ErrInvalidInput
	}

	product, err := s.repo.Create(ctx, &clean)
	if err != nil {
		return nil, normalizeError(err)
	}
	return product, nil
}

// GetByID – получение продукта по ID
func (s *productService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if blank(id) {
		return nil, ErrInvalidInput
	}
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, normalizeError(err)
	}
	return product, nil
}

// GetByCustomerID – получение всех продуктов покупателя
func (s *productService) GetByCustomerID(ctx context.Context, customerID string) ([]domain.Product, error) {
	if blank(customerID) {
		return nil, ErrInvalidInput
	}
	products, err := s.repo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, normalizeError(err)
	}
	return products, nil
}

// Update – частичное обновление продукта
func (s *productService) Update(ctx context.Context, id string, dto *domain.UpdateProductDTO) (*domain.Product, error) {
	if blank(id) || dto == nil {
		return nil, ErrInvalidInput
	}
	if dto.Status != nil {
		if !isValidProductStatus(*dto.Status) {
			return nil, ErrInvalidInput
		}
		// Не разрешаем устанавливать статус archived через этот метод (используйте Delete)
		if *dto.Status == domain.ProductArchived {
			return nil, ErrInvalidInput
		}
	}
	// Очистка строковых полей
	if dto.Title != nil {
		trimmed := strings.TrimSpace(*dto.Title)
		dto.Title = &trimmed
	}
	if dto.Description != nil {
		trimmed := strings.TrimSpace(*dto.Description)
		dto.Description = &trimmed
	}
	if dto.Image != nil {
		trimmed := strings.TrimSpace(*dto.Image)
		dto.Image = &trimmed
	}
	if dto.Location != nil {
		trimmed := strings.TrimSpace(*dto.Location)
		dto.Location = &trimmed
	}

	product, err := s.repo.Update(ctx, id, dto)
	if err != nil {
		return nil, normalizeError(err)
	}
	return product, nil
}

// Delete – мягкое удаление (устанавливает статус archived)
func (s *productService) Delete(ctx context.Context, id string) error {
	if blank(id) {
		return ErrInvalidInput
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return normalizeError(err)
	}
	return nil
}

// List – список продуктов с пагинацией
func (s *productService) List(ctx context.Context, offset, limit int) ([]domain.Product, error) {
	o, l, err := validatePage(offset, limit)
	if err != nil {
		return nil, err
	}
	products, err := s.repo.List(ctx, o, l)
	if err != nil {
		return nil, normalizeError(err)
	}
	return products, nil
}

// Search – полнотекстовый поиск
func (s *productService) Search(ctx context.Context, q string, categoryID *string) ([]domain.Product, error) {
	q = strings.TrimSpace(q)
	if q == "" && (categoryID == nil || *categoryID == "") {
		return nil, ErrInvalidInput
	}
	if categoryID != nil && *categoryID == "" {
		categoryID = nil
	}
	products, err := s.repo.Search(ctx, q, categoryID)
	if err != nil {
		return nil, normalizeError(err)
	}
	return products, nil
}

// GetExchangeCandidates – кандидаты для обмена
func (s *productService) GetExchangeCandidates(ctx context.Context, productID string) ([]domain.Product, error) {
	if blank(productID) {
		return nil, ErrInvalidInput
	}
	products, err := s.repo.GetExchangeCandidates(ctx, productID)
	if err != nil {
		return nil, normalizeError(err)
	}
	return products, nil
}

// Вспомогательные функции
func isValidProductStatus(status domain.ProductStatus) bool {
	switch status {
	case domain.ProductActive, domain.ProductReserved, domain.ProductExchanged, domain.ProductArchived:
		return true
	}
	return false
}

// blank – проверка на пустую строку
func blank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// validatePage – нормализация offset/limit
func validatePage(offset, limit int) (int, int, error) {
	if offset < 0 {
		offset = 0
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return offset, limit, nil
}

// normalizeError – преобразование ошибок репозитория в доменные ошибки
func normalizeError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	// Если ошибка уже является одной из наших (ErrInvalidInput, ErrNotFound), возвращаем как есть
	if errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrNotFound) {
		return err
	}
	// Все остальные ошибки считаем внутренними. Оригинал оборачивается, а не
	// отбрасывается: errors.Is(err, ErrInternal) продолжает работать, но
	// причина доезжает до лога — без неё отладка 500-х сводится к угадыванию.
	return fmt.Errorf("%w: %v", ErrInternal, err)
}

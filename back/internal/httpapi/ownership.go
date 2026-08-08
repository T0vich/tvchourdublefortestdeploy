package httpapi

import (
	"net/http"
	"trade-chain/internal/auth"
	"trade-chain/internal/service"
)

// Проверки владельца намеренно живут на транспортном слое, а не в сервисах.
//
// Идентификатор пользователя приходит из токена, а токен разбирает middleware —
// то есть здесь. Сервисные интерфейсы принимают только идентификатор объекта и
// не знают, кто именно выполняет действие; чтобы научить их этому, пришлось бы
// менять сигнатуры во всех сервисах и репозиториях сразу.
//
// Цена решения — один дополнительный запрос на чтение перед изменением.

// actor достаёт идентификатор пользователя из токена.
// При отсутствии отвечает 401 и возвращает false.
func actor(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeError(w, service.ErrUnauthorized)
		return "", false
	}

	return userID, true
}

// requireSelf разрешает действие только над собственным аккаунтом.
func requireSelf(w http.ResponseWriter, r *http.Request, targetCustomerID string) bool {
	userID, ok := actor(w, r)
	if !ok {
		return false
	}

	if userID != targetCustomerID {
		writeError(w, service.ErrForbidden)
		return false
	}

	return true
}

// requireProductOwner разрешает действие только владельцу товара.
func requireProductOwner(
	w http.ResponseWriter,
	r *http.Request,
	products service.ProductService,
	productID string,
) bool {
	userID, ok := actor(w, r)
	if !ok {
		return false
	}

	product, err := products.GetByID(r.Context(), productID)
	if err != nil {
		writeError(w, err)
		return false
	}

	if product.CustomerID != userID {
		writeError(w, service.ErrForbidden)
		return false
	}

	return true
}

// requireWishlistOwner разрешает действие только владельцу товара, к которому
// привязан список желаний.
func requireWishlistOwner(
	w http.ResponseWriter,
	r *http.Request,
	wishlists service.WishlistService,
	products service.ProductService,
	wishlistID string,
) bool {
	userID, ok := actor(w, r)
	if !ok {
		return false
	}

	wishlist, err := wishlists.GetByID(r.Context(), wishlistID)
	if err != nil {
		writeError(w, err)
		return false
	}

	product, err := products.GetByID(r.Context(), wishlist.ProductID)
	if err != nil {
		writeError(w, err)
		return false
	}

	if product.CustomerID != userID {
		writeError(w, service.ErrForbidden)
		return false
	}

	return true
}

// requireReviewAuthor разрешает удаление отзыва только его автору.
func requireReviewAuthor(
	w http.ResponseWriter,
	r *http.Request,
	reviews service.ReviewService,
	reviewID string,
) bool {
	userID, ok := actor(w, r)
	if !ok {
		return false
	}

	review, err := reviews.GetByID(r.Context(), reviewID)
	if err != nil {
		writeError(w, err)
		return false
	}

	if review.FromCustomerID != userID {
		writeError(w, service.ErrForbidden)
		return false
	}

	return true
}

// requireChainInitiator разрешает действие только инициатору цепочки.
func requireChainInitiator(
	w http.ResponseWriter,
	r *http.Request,
	chains service.ChainService,
	chainID string,
) bool {
	userID, ok := actor(w, r)
	if !ok {
		return false
	}

	chain, err := chains.GetByID(r.Context(), chainID)
	if err != nil {
		writeError(w, err)
		return false
	}

	if chain.InitiatorID != userID {
		writeError(w, service.ErrForbidden)
		return false
	}

	return true
}

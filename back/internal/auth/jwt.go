package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// devSecret используется только когда JWT_SECRET не задан — то есть при
// локальном запуске. В любом развёрнутом окружении секрет обязан приходить
// из переменной окружения, иначе подписать чужой токен может кто угодно,
// у кого есть доступ к исходникам.
const devSecret = "local-development-only-do-not-deploy"

var jwtSecret = loadSecret()

func loadSecret() []byte {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}

	log.Printf("ВНИМАНИЕ: JWT_SECRET не задан, используется отладочный секрет")

	return []byte(devSecret)
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	// WithValidMethods обязателен: без него токен, подписанный другим
	// алгоритмом (или вовсе не подписанный), дошёл бы до проверки подписи
	// и мог быть принят.
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

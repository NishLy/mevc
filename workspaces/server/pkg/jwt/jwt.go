package pkg

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func VerifyToken(tokenStr, secret, tokenType string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return "", errors.New("TOKEN_EXPIRED")
			}
		}
		return "", errors.New("INVALID_TOKEN")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("INVALID_TOKEN_CLAIMS")
	}

	jwtType, ok := claims["type"].(string)
	if !ok || jwtType != tokenType {
		return "", errors.New("INVALID_TOKEN_TYPE")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("INVALID_TOKEN_SUB")
	}

	return userID, nil
}

func GenerateToken(userID, secret, tokenType string, expiresIn int64) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": tokenType,
		"exp":  time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

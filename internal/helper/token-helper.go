package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ExtractJwtClaim[T any](jwtString string, secret string) (T, error) {
	var zero T

	jwtString = strings.TrimSpace(jwtString)
	if jwtString == "" {
		return zero, errors.New("jwt string is empty")
	}
	if secret == "" {
		return zero, errors.New("secret is empty")
	}

	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS512"}))

	if err != nil {
		return zero, fmt.Errorf("failed to parse jwt: %w", err)
	}

	if !token.Valid {
		return zero, errors.New("invalid or expired jwt token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return zero, errors.New("invalid jwt claims format")
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return zero, fmt.Errorf("failed to marshal claims: %w", err)
	}

	var result T
	if err := json.Unmarshal(claimsBytes, &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal claim into target type: %w", err)
	}

	return result, nil
}

func SignJwt[T any](claims T, secret string, expiresIn time.Duration) (string, error) {
	claimsMap := make(jwt.MapClaims)

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	if err := json.Unmarshal(claimsBytes, &claimsMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal claims into map: %w", err)
	}

	now := time.Now().Unix()
	if _, exists := claimsMap["iat"]; !exists {
		claimsMap["iat"] = now
	}
	if _, exists := claimsMap["exp"]; !exists && expiresIn > 0 {
		claimsMap["exp"] = now + int64(expiresIn.Seconds())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claimsMap)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

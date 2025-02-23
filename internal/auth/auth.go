package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	cost := 10 // iteration of encryption
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// return nil if password is correct
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

// JWT is a way to authenticate user after log in (in the session).
// It store users' data. Server issues a JWT, sign it (use token secret
// to convert json to long string) and send back to client.
// NOTE: signing is NOT encoding. People can read JWT, but only
// server, that has token secret, can validate it.
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	method := jwt.SigningMethodHS256
	claim := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(), // easy way to uuid -> string
	}
	token := jwt.NewWithClaims(method, claim)

	signedString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedString, nil
}

// NOTE: only server can validate the JWT because only server knows
// token secret (which used to sign the string)
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimStruct := jwt.RegisteredClaims{}
	// keyFunc will check if the token is valid (too lazy, not do it)
	// by intially parsed token, we have method + header + claims to play with
	// if not valid -> return error, if valid -> return token secret
	// ParseWithClaims() needs that secret to parse the other part of token
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		// return []byte because HS256 belong to HMAC signing
		// use []byte to verify key
		return []byte(tokenSecret), nil
	}
	parsedToken, err := jwt.ParseWithClaims(tokenString, &claimStruct, keyFunc)
	if err != nil {
		return uuid.UUID{}, err
	}

	userID, err := parsedToken.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userUUID, nil
}

// extract the token string from header in form "Bearer <token>"
func GetBearerToken(header http.Header) (string, error) {
	// it should be in this form "Bearer TOKEN_STRING"
	authorizationHeader := header.Get("Authorization")

	tokenString, ok := strings.CutPrefix(authorizationHeader, "Bearer ")
	if !ok {
		return "", fmt.Errorf("token bearer not found")
	}

	return tokenString, nil
}

// generate random 256-bit string data
func MakeRefreshToken() (string, error) {
	randomData := make([]byte, 32)
	_, err := rand.Read(randomData)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(randomData)
	return token, nil
}

// extract api-key from "Authorization" header
// in "Api <key>" form
func GetAPIKey(headers http.Header) (string, error) {
	authorizationHeader := headers.Get("Authorization")
	apiKey, ok := strings.CutPrefix(authorizationHeader, "ApiKey ")
	if !ok {
		return "", fmt.Errorf("Not found api key")
	}

	return apiKey, nil
}

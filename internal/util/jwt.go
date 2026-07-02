package util

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 💡 Ganti dengan kata rahasia lu sendiri yang super aman, amankan di .env jika perlu
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Custom Claims untuk menyimpan data user di dalam JWT
type JWTClaims struct {
	GoogleID string `json:"google_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// 1. Fungsi untuk membuat Token JWT (Berlaku 24 Jam)
func GenerateJWT(googleID, name, email string) (string, error) {
	claims := JWTClaims{
		GoogleID: googleID,
		Name:     name,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 2. Fungsi untuk memvalidasi dan membongkar isi Token JWT
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Pastikan algoritma pengunciannya sesuai (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("algoritma token tidak valid")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Ambil data claims-nya jika token valid
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token tidak valid atau kadaluwarsa")
}
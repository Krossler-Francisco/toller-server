package auth

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Middleware para verificar JWT
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener el token del header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "unauthorized", "message": "Token no proporcionado"}`, http.StatusUnauthorized)
			return
		}

		// Formato esperado: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "unauthorized", "message": "Formato de token inválido"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Verificar y parsear el token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verificar que el método de firma sea el esperado
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error": "unauthorized", "message": "Token inválido o expirado"}`, http.StatusUnauthorized)
			return
		}

		// Extraer los claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "unauthorized", "message": "Claims inválidos"}`, http.StatusUnauthorized)
			return
		}

		// Obtener el user_id del token
		userID, ok := claims["user_id"].(float64) // JWT parsea números como float64
		if !ok {
			http.Error(w, `{"error": "unauthorized", "message": "User ID no encontrado en el token"}`, http.StatusUnauthorized)
			return
		}

		// Agregar el user_id al contexto de la request
		ctx := context.WithValue(r.Context(), "user_id", int(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

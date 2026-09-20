package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"banking-system/internal/auth"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "authorization token is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "authorization header format must be Bearer <token>"})
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid or expired token"})
			return
		}

		ctx := auth.SetCustomerIDInContext(r.Context(), claims.CustomerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

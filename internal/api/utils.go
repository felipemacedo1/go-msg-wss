package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"

	"github.com/felipemacedo1/go-msg-wss/internal/store/pgstore"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (h apiHandler) readRoom(
	w http.ResponseWriter,
	r *http.Request,
) (room pgstore.Room, rawRoomID string, roomID uuid.UUID, ok bool) {
	rawRoomID = chi.URLParam(r, "room_id")
	roomID, err := uuid.Parse(rawRoomID)
	if err != nil {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return pgstore.Room{}, "", uuid.UUID{}, false
	}

	room, err = h.q.GetRoom(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusBadRequest)
			return pgstore.Room{}, "", uuid.UUID{}, false
		}

		slog.Error("failed to get room", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return pgstore.Room{}, "", uuid.UUID{}, false
	}

	return room, rawRoomID, roomID, true
}

func sendJSON(w http.ResponseWriter, rawData any) {
	data, _ := json.Marshal(rawData)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// GetJWTSecret retorna a chave JWT do ambiente
func GetJWTSecret() []byte {
	jwtKey := getEnv("MSGWSS_JWT_SECRET", "")
	if jwtKey == "" {
		if os.Getenv("GO_ENV") != "test" {
			slog.Error("MSGWSS_JWT_SECRET não configurado. Este é um requisito para a segurança da aplicação!")
			os.Exit(1)
		}
		return []byte("test-jwt-secret")
	}
	return []byte(jwtKey)
}

// getEnv obtém variável de ambiente com fallback para valor padrão
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// Lazy-loaded JWT secret
var jwtSecret []byte
var jwtSecretOnce sync.Once

// GetJWTSecretLazy obtém a chave JWT das variáveis de ambiente com inicialização preguiçosa
func GetJWTSecretLazy() []byte {
	jwtSecretOnce.Do(func() {
		jwtSecret = GetJWTSecret()
	})
	return jwtSecret
}

func extractClaimsFromJWT(r *http.Request) map[string]interface{} {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil
	}

	tokenString := parts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Se seu AuthService usar RS256, aqui você precisa carregar a public key
		// Se for HS256 (ex: com guest simples), você pode usar o jwtSecret
		return GetJWTSecretLazy(), nil
	})

	if err != nil || !token.Valid {
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims
	}

	return nil
}

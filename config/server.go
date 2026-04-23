package config

import (
	"net/http"

	"github.com/assidik12/catalyst/internal/delivery/http/middleware"
)

func NewServer(authMiddleware *middleware.AuthMiddleware) *http.Server {
	return &http.Server{
		Addr:    ":" + GetConfig().AppPort,
		Handler: authMiddleware,
	}
}

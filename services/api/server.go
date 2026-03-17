package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Hunsin/compass/lib/auth"
)

// Server represents the HTTP JSON API server
type Server struct {
	mux       *http.ServeMux
	kcClient  auth.KeycloakClient
	validator auth.JWTValidator
}

// NewServer creates a new API server
func NewServer(kcClient auth.KeycloakClient, validator auth.JWTValidator) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		kcClient:  kcClient,
		validator: validator,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Public routes
	s.mux.HandleFunc("POST /api/login", s.handleLogin)

	// Protected routes
	s.mux.Handle("GET /api/me", auth.HTTPMiddleware(s.validator)(http.HandlerFunc(s.handleMe)))
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password carry required", http.StatusBadRequest)
		return
	}

	// Call Keycloak Token endpoint via password grant
	tokenResp, err := s.kcClient.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		log.Printf("Login failed for user %s: %v", req.Username, err)
		http.Error(w, "Invalid credentials or upstream error", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokenResp); err != nil {
		log.Printf("Failed to encode token response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID,
	})
}

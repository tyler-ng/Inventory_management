package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/inventory-management-system/internal/models"
)

// generateJWT generates a secure JWT token
func generateJWT(user models.User) (string, error) {
	// Get JWT secret from environment with fallback
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("your-secure-fallback-secret")
	}

	// Create token with enhanced claims
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // 24-hour expiration
		"iat":        time.Now().Unix(),
		"nbf":        time.Now().Unix(), // Not before current time
		"token_type": "access",
	}

	// Create token with strong signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	return token.SignedString(jwtSecret)
}

// Additional security enhancements in login handler
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	
	// Decode request body with size limit
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit
	
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if request.Username == "" || request.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	
	// Find user by username
	var user models.User
	if err := h.db.Where("username = ? AND status = 'active'", request.Username).First(&user).Error; err != nil {
		// Deliberate delay to prevent timing attacks
		time.Sleep(time.Second)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Constant-time password comparison
	if !user.CheckPassword(request.Password) {
		// Deliberate delay to prevent timing attacks
		time.Sleep(time.Second)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Generate JWT token
	token, err := generateJWT(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	
	// Update last login time
	h.db.Model(&user).Update("last_login", time.Now())
	
	// Clear sensitive fields
	user.PasswordHash = ""
	
	// Return response
	response := LoginResponse{
		Token: token,
		User:  user,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
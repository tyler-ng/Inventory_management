package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/inventory-management-system/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Token  string       `json:"token"`
	User   models.User  `json:"user"`
}

// RegisterRequest represents the register request payload
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// Login handles user authentication and returns JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	
	// Decode request body
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
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Check password
	if !user.CheckPassword(request.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	// Update last login time
	user.LastLogin = time.Now()
	h.db.Save(&user)
	
	// Generate JWT token
	token, err := generateJWT(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	
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

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	
	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if request.Username == "" || request.Password == "" || request.Email == "" || request.FullName == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}
	
	// Check if username already exists
	var existingUser models.User
	if err := h.db.Where("username = ?", request.Username).First(&existingUser).Error; err == nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}
	
	// Check if email already exists
	if err := h.db.Where("email = ?", request.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}
	
	// Create new user
	user := models.User{
		Username: request.Username,
		Email:    request.Email,
		FullName: request.FullName,
		Role:     "user", // Default role
		Status:   "active",
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.PasswordHash = string(hashedPassword)
	
	// Save user
	if err := h.db.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Generate JWT token
	token, err := generateJWT(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	
	// Clear sensitive fields
	user.PasswordHash = ""
	
	// Return response
	response := LoginResponse{
		Token: token,
		User:  user,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// generateJWT generates a JWT token for the user
func generateJWT(user models.User) (string, error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"username": user.Username,
		"email": user.Email,
		"role": user.Role,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiration
	}
	
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Get JWT secret from environment
	jwtSecret := []byte("your-secret-key") // This should be loaded from config in a real app
	
	// Generate encoded token
	return token.SignedString(jwtSecret)
}
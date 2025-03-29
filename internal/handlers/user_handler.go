package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yourusername/inventory-management-system/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserHandler handles HTTP requests for user endpoints
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GetUsers handles GET requests to retrieve all users
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	
	// Apply filters if any
	query := h.db
	
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	
	if role := r.URL.Query().Get("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	
	if search := r.URL.Query().Get("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?", 
			searchPattern, searchPattern, searchPattern)
	}
	
	// Apply pagination
	page := 1
	limit := 10
	
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 {
			limit = limitNum
		}
	}
	
	offset := (page - 1) * limit
	
	// Execute query and omit password hash
	if err := query.Select("id, username, email, full_name, role, status, last_login, created_at, updated_at").
		Order("username ASC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		http.Error(w, "Failed to retrieve users: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetUser handles GET requests to retrieve a single user
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	var user models.User
	if err := h.db.Select("id, username, email, full_name, role, status, last_login, created_at, updated_at").
		First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// CreateUser handles POST requests to create a new user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if user.Username == "" || user.Email == "" || user.FullName == "" || user.PasswordHash == "" {
		http.Error(w, "Username, email, full name, and password are required", http.StatusBadRequest)
		return
	}
	
	// Check if username already exists
	var existingUser models.User
	if err := h.db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}
	
	// Check if email already exists
	if err := h.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "Email already in use", http.StatusConflict)
		return
	}
	
	// Set default role and status if not provided
	if user.Role == "" {
		user.Role = "user"
	}
	
	if user.Status == "" {
		user.Status = "active"
	}
	
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password: "+err.Error(), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = string(hashedPassword)
	
	// Create user in database
	if err := h.db.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Remove password hash from response
	user.PasswordHash = ""
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// UpdateUser handles PUT requests to update an existing user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	// Check if user exists
	var existingUser models.User
	if err := h.db.First(&existingUser, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Parse request body
	var updatedUser models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Set the ID to ensure we're updating the correct record
	updatedUser.ID = uint(id)
	
	// Check if username is being changed and if it's already taken
	if updatedUser.Username != "" && updatedUser.Username != existingUser.Username {
		var count int64
		if err := h.db.Model(&models.User{}).Where("username = ? AND id != ?", updatedUser.Username, id).Count(&count).Error; err != nil {
			http.Error(w, "Failed to check username uniqueness: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		if count > 0 {
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}
	}
	
	// Check if email is being changed and if it's already in use
	if updatedUser.Email != "" && updatedUser.Email != existingUser.Email {
		var count int64
		if err := h.db.Model(&models.User{}).Where("email = ? AND id != ?", updatedUser.Email, id).Count(&count).Error; err != nil {
			http.Error(w, "Failed to check email uniqueness: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		if count > 0 {
			http.Error(w, "Email already in use", http.StatusConflict)
			return
		}
	}
	
	// Don't update password with this endpoint
	updatedUser.PasswordHash = existingUser.PasswordHash
	
	// Update in database
	if err := h.db.Model(&updatedUser).Select("username", "email", "full_name", "role", "status").Updates(updatedUser).Error; err != nil {
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Retrieve updated user
	var finalUser models.User
	if err := h.db.Select("id, username, email, full_name, role, status, last_login, created_at, updated_at").
		First(&finalUser, id).Error; err != nil {
		http.Error(w, "Failed to retrieve updated user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalUser)
}

// DeleteUser handles DELETE requests to delete a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	// Check if user exists
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Check if user is the last admin
	if user.Role == "admin" {
		var adminCount int64
		if err := h.db.Model(&models.User{}).Where("role = ? AND id != ?", "admin", id).Count(&adminCount).Error; err != nil {
			http.Error(w, "Failed to check admin count: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		if adminCount == 0 {
			http.Error(w, "Cannot delete the last admin user", http.StatusBadRequest)
			return
		}
	}
	
	// Soft delete by updating status
	if err := h.db.Model(&user).Update("status", "inactive").Error; err != nil {
		http.Error(w, "Failed to deactivate user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// GetCurrentUser handles GET requests to retrieve the current authenticated user
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	
	var user models.User
	if err := h.db.Select("id, username, email, full_name, role, status, last_login, created_at, updated_at").
		First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ChangePassword handles POST requests to change a user's password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	
	// Parse request body
	var request struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate request
	if request.CurrentPassword == "" || request.NewPassword == "" {
		http.Error(w, "Current password and new password are required", http.StatusBadRequest)
		return
	}
	
	// Get user with password hash
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.CurrentPassword)); err != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}
	
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update password in database
	if err := h.db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		http.Error(w, "Failed to update password: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
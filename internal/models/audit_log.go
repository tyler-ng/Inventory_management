package models

import (
	"time"

	"gorm.io/gorm"
)

// AuditLog represents a record of user actions in the system
type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	Action     string    `json:"action" gorm:"not null"`
	EntityType string    `json:"entity_type" gorm:"not null"`
	EntityID   uint      `json:"entity_id" gorm:"not null"`
	OldValues  string    `json:"old_values" gorm:"type:jsonb"`
	NewValues  string    `json:"new_values" gorm:"type:jsonb"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// Relationships
	User       *User     `json:"user" gorm:"foreignKey:UserID"`
}

// CreateAuditLog creates a new audit log entry
func CreateAuditLog(db *gorm.DB, userID uint, action, entityType string, entityID uint, oldValues, newValues, ipAddress string) error {
	auditLog := AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValues:  oldValues,
		NewValues:  newValues,
		IPAddress:  ipAddress,
	}
	
	return db.Create(&auditLog).Error
}
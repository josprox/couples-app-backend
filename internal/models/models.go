package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base model struct adding UUID to all tables natively
type Base struct {
	ID        string         `gorm:"type:varchar(36);primary_key;" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (base *Base) BeforeCreate(tx *gorm.DB) error {
	base.ID = uuid.New().String()
	return nil
}

type User struct {
	Base
	Name               string    `gorm:"type:varchar(100);not null" json:"name"`
	Email              string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash       string    `gorm:"type:varchar(255);not null" json:"-"`
	BirthDate          time.Time `gorm:"type:date;not null" json:"birth_date"`
	ProfilePictureURL  string    `gorm:"type:varchar(500)" json:"profile_picture_url"`
	ZodiacSign         string    `gorm:"type:varchar(20)" json:"zodiac_sign"`
	FCMToken           string    `gorm:"type:varchar(500)" json:"fcm_token"`
	RelationshipID     string    `gorm:"type:varchar(36)" json:"relationship_id"`
}

type Relationship struct {
	Base
	User1ID    string `gorm:"type:varchar(36);not null;uniqueIndex:idx_user1_rel" json:"user1_id"`
	User2ID    string `gorm:"type:varchar(36);not null;uniqueIndex:idx_user2_rel" json:"user2_id"`
	StartDate  time.Time `gorm:"type:date" json:"start_date"`
	HowWeMet   string `gorm:"type:text" json:"how_we_met"`
	
	// Relations (Foreign Keys)
	User1 User `gorm:"foreignKey:User1ID" json:"user1"`
	User2 User `gorm:"foreignKey:User2ID" json:"user2"`
}

type RelationshipRequest struct {
	Base
	SenderID   string `gorm:"type:varchar(36);not null" json:"sender_id"`
	ReceiverID string `gorm:"type:varchar(36);not null" json:"receiver_id"`
	Status     string `gorm:"type:enum('pending', 'accepted', 'rejected');default:'pending'" json:"status"`
	
	Sender   User `gorm:"foreignKey:SenderID" json:"sender"`
	Receiver User `gorm:"foreignKey:ReceiverID" json:"receiver"`
}

type Message struct {
	Base
	RelationshipID string `gorm:"type:varchar(36);not null;index:idx_messages_relationship_time,priority:1" json:"relationship_id"`
	SenderID       *string `gorm:"type:varchar(36)" json:"sender_id"` // NULL if AI
	MessageType    string `gorm:"type:enum('text', 'image', 'location', 'ai');default:'text'" json:"message_type"`
	Content        string `gorm:"type:text;not null" json:"content"`
	Status         string `gorm:"type:enum('sent', 'delivered', 'read');default:'sent'" json:"status"`
	Metadata       string `gorm:"type:json" json:"metadata"` // Storing JSON string or mapped to a map
	
	// Custom Created At index for pagination
	CreatedAt time.Time `gorm:"index:idx_messages_relationship_time,priority:2;index:idx_created_at" json:"created_at"`

	Relationship Relationship `gorm:"foreignKey:RelationshipID" json:"-"`
	Sender       *User        `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

type Location struct {
	Base
	UserID         string  `gorm:"type:varchar(36);not null;index:idx_location_user_time,priority:1" json:"user_id"`
	RelationshipID string  `gorm:"type:varchar(36);not null" json:"relationship_id"`
	Latitude       float64 `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude      float64 `gorm:"type:decimal(11,8);not null" json:"longitude"`
	Timestamp      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index:idx_location_user_time,priority:2" json:"timestamp"`

	User         User         `gorm:"foreignKey:UserID" json:"-"`
	Relationship Relationship `gorm:"foreignKey:RelationshipID" json:"-"`
}

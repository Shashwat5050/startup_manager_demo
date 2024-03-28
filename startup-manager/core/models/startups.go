package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// StartupInfo represents the startups_info table in the database.
type StartupInfo struct {
	ID            	uuid.UUID              `json:"id" db:"id"`
	ServerID      	uuid.UUID              `json:"server_id" db:"server_id"`
	Variables     	map[string]interface{} `json:"variables" db:"variables"`
	StartupCommand  string					`json:"startup_command" db:"startup_command`
	CreatedAt     	time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     	*time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt     	*time.Time              `json:"deleted_at" db:"deleted_at"`
}

// JSONB represents a JSONB data type.
// You might need to implement this type based on your database driver.
type JSONB map[string]interface{}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
    // Implement the logic to convert the value from the database to your JSONB type.
    // For example, you can use json.Unmarshal.
    return json.Unmarshal(value.([]byte), j)
}

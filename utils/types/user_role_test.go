package types

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_NewUserRole(t *testing.T) {
	tests := []struct {
		input    string
		expected UserRole
	}{
		{"admin", UserRoleAdmin},
		{"user", UserRoleUser},
		{"invalid", UserRoleUser}, // Defaults to UserRoleUser
		{"", UserRoleUser},        // Defaults to UserRoleUser
		{"ADMIN", UserRoleUser},   // Case sensitive, defaults to UserRoleUser
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NewUserRole(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_String(t *testing.T) {
	assert.Equal(t, "admin", UserRoleAdmin.String())
	assert.Equal(t, "user", UserRoleUser.String())
	assert.Equal(t, "invalid", UserRole("invalid").String())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		role     UserRole
		expected bool
	}{
		{UserRoleAdmin, true},
		{UserRoleUser, true},
		{UserRole("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.IsValid())
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_MarshalJSON(t *testing.T) {
	tests := []struct {
		role     UserRole
		expected string
		hasError bool
	}{
		{UserRoleAdmin, `"admin"`, false},
		{UserRoleUser, `"user"`, false},
		{UserRole("invalid"), "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			data, err := json.Marshal(tt.role)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(data))
			}
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		data     string
		expected UserRole
		hasError bool
	}{
		{`"admin"`, UserRoleAdmin, false},
		{`"user"`, UserRoleUser, false},
		{`"invalid"`, "", true},
		// Invalid JSON cases
		{`123`, "", true},  // Number instead of string
		{`true`, "", true}, // Boolean instead of string
		{`null`, "", true}, // Null value
		{`{`, "", true},    // Malformed JSON
		{`[`, "", true},    // Malformed JSON array
		{`"`, "", true},    // Incomplete string
	}

	for _, tt := range tests {
		t.Run(tt.data, func(t *testing.T) {
			var role UserRole
			err := json.Unmarshal([]byte(tt.data), &role)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, role)
			}
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected UserRole
		hasError bool
	}{
		{"valid admin", "admin", UserRoleAdmin, false},
		{"valid user", "user", UserRoleUser, false},
		{"invalid role", "invalid", "", true},
		// Non-string input cases
		{"nil input", nil, "", true},
		{"int input", 123, "", true},
		{"bool input", true, "", true},
		{"float input", 123.45, "", true},
		{"byte slice", []byte("admin"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var role UserRole
			err := role.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
				if tt.input != nil {
					// Check error message for type assertion failure
					if _, ok := tt.input.(string); !ok {
						assert.Contains(t, err.Error(), "invalid data type")
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, role)
			}
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_Value(t *testing.T) {
	tests := []struct {
		role     UserRole
		expected driver.Value
		hasError bool
	}{
		{UserRoleAdmin, "admin", false},
		{UserRoleUser, "user", false},
		{UserRole("invalid"), nil, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			value, err := tt.role.Value()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, value)
			}
		})
	}
}

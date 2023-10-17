package base

import (
	"testing"
	"time"
)

func TestTimeFields_SetTimes(t *testing.T) {
	tests := []struct {
		name string
		att  *TimeFields
	}{
		{
			name: "Test SetTimes",
			att:  &TimeFields{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.att.SetTimes()
			currentTime := time.Now()
			if tt.att.CreatedAt == nil || tt.att.UpdatedAt == nil {
				t.Errorf("Expected CreatedAt and UpdatedAt to be set, but one or both are nil")
			}
			if !tt.att.CreatedAt.Equal(currentTime) || !tt.att.UpdatedAt.Equal(currentTime) {
				t.Errorf("Expected CreatedAt and UpdatedAt to be equal to the current time")
			}
		})
	}
}

func TestTimeFields_SetUpdateTime(t *testing.T) {
	tests := []struct {
		name string
		att  *TimeFields
	}{
		{
			name: "Test SetUpdateTime",
			att:  &TimeFields{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.att.SetUpdateTime()
			currentTime := time.Now()
			if tt.att.UpdatedAt == nil {
				t.Errorf("Expected UpdatedAt to be set, but it is nil")
			}
			if !tt.att.UpdatedAt.Equal(currentTime) {
				t.Errorf("Expected UpdatedAt to be equal to the current time")
			}
		})
	}
}

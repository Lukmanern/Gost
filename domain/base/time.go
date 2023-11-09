package base

import "time"

// TimeFields struct used by almost all entity.
// This struct give stabillity for creating more and more entity/ struct/ table.
// This struct prevents developers from making typing errors / typo.
type TimeFields struct {
	CreatedAt *time.Time `gorm:"type:timestamp null;default:null" json:"created_at"`
	UpdatedAt *time.Time `gorm:"type:timestamp null;default:null" json:"updated_at"`
}

func (att *TimeFields) SetCreateTimes() {
	timeNow := time.Now()
	att.CreatedAt = &timeNow
	att.UpdatedAt = &timeNow
}

func (att *TimeFields) SetUpdateTime() {
	timeNow := time.Now()
	att.UpdatedAt = &timeNow
}

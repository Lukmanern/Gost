package base

import "time"

type TimeAttributes struct {
	CreatedAt *time.Time `gorm:"type:timestamp null;default:null" json:"created_at"`
	UpdatedAt *time.Time `gorm:"type:timestamp null;default:null" json:"updated_at"`
}

func (att *TimeAttributes) SetTimeAttributes() {
	timeNow := time.Now()
	att.CreatedAt = &timeNow
	att.UpdatedAt = &timeNow
}

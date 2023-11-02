package base

import "time"

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

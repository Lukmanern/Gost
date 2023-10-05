package base

import "time"

type TimeFieds struct {
	CreatedAt *time.Time `gorm:"type:timestamp null;default:null" json:"created_at"`
	UpdatedAt *time.Time `gorm:"type:timestamp null;default:null" json:"updated_at"`
}

func (att *TimeFieds) SetTimes() {
	timeNow := time.Now()
	att.CreatedAt = &timeNow
	att.UpdatedAt = &timeNow
}

func (att *TimeFieds) SetUpdateTime() {
	timeNow := time.Now()
	att.CreatedAt = &timeNow
	att.UpdatedAt = &timeNow
}

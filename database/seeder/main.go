package seeder

import "gorm.io/gorm"

func Seed(db *gorm.DB, data interface{}) error {
	if err := db.Create(data).Error; err != nil {
		return err
	}

	return nil
}

// ⚠️ Don't forget to Add Your new Table
// at AllTables func at all_entities.go file

package entity

type Role struct {
	ID          int    `gorm:"type:serial;primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(255) not null unique" json:"name"`
	Description string `gorm:"type:varchar(255) not null" json:"description"`
	TimeFields
}

func (e *Role) TableName() string {
	return "roles"
}

package repository

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"gorm.io/gorm"
)

type RoleRepository interface {
	// Create adds a new role to the repository.
	Create(ctx context.Context, role entity.Role) (id int, err error)

	// GetByID retrieves a role by its unique identifier.
	GetByID(ctx context.Context, id int) (role *entity.Role, err error)

	// GetByName retrieves a role by its name.
	GetByName(ctx context.Context, name string) (role *entity.Role, err error)

	// GetAll retrieves all roles based on a filter for pagination.
	GetAll(ctx context.Context, filter model.RequestGetAll) (roles []entity.Role, total int, err error)

	// Update modifies role information in the repository.
	Update(ctx context.Context, role entity.Role) (err error)

	// Delete removes a role from the repository by its ID.
	Delete(ctx context.Context, id int) (err error)
}

type RoleRepositoryImpl struct {
	db *gorm.DB
}

var (
	roleRepositoryImpl     *RoleRepositoryImpl
	roleRepositoryImplOnce sync.Once
)

func NewRoleRepository() RoleRepository {
	roleRepositoryImplOnce.Do(func() {
		roleRepositoryImpl = &RoleRepositoryImpl{
			db: connector.LoadDatabase(),
		}
	})
	return roleRepositoryImpl
}

func (repo *RoleRepositoryImpl) Create(ctx context.Context, role entity.Role) (id int, err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&role)
		if res.Error != nil {
			tx.Rollback()
			return res.Error
		}
		id = role.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *RoleRepositoryImpl) GetByID(ctx context.Context, id int) (role *entity.Role, err error) {
	role = &entity.Role{}
	result := repo.db.Where("id = ?", id).First(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role, nil
}

func (repo *RoleRepositoryImpl) GetByName(ctx context.Context, name string) (role *entity.Role, err error) {
	role = &entity.Role{}
	result := repo.db.Where("name = ?", name).First(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role, nil
}

func (repo *RoleRepositoryImpl) GetAll(ctx context.Context, filter model.RequestGetAll) (roles []entity.Role, total int, err error) {
	var count int64
	filter.Sort = ""
	args := []interface{}{"%" + filter.Keyword + "%"}
	cond := "name LIKE ?"
	result := repo.db.Where(cond, args...).Find(&roles)
	count = result.RowsAffected
	if result.Error != nil {
		return nil, 0, result.Error
	}
	roles = []entity.Role{}
	skip := int64(filter.Limit * (filter.Page - 1))
	limit := int64(filter.Limit)
	result = repo.db.Where(cond, args...).Limit(int(limit)).Offset(int(skip)).Find(&roles)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	total = int(count)
	return roles, total, nil
}

func (repo *RoleRepositoryImpl) Update(ctx context.Context, role entity.Role) (err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		var oldData entity.Role
		result := tx.Where("id = ?", role.ID).First(&oldData)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		oldData.Name = role.Name
		oldData.Description = role.Description
		oldData.UpdatedAt = role.UpdatedAt
		result = tx.Save(&oldData)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
		return nil
	})

	return err
}

func (repo *RoleRepositoryImpl) Delete(ctx context.Context, id int) (err error) {
	deleted := entity.Role{}
	result := repo.db.Where("id = ?", id).Delete(&deleted)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

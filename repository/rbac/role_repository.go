package repository

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role entity.Role, permissionsID []int) (id int, err error)
	GetByID(ctx context.Context, id int) (role *entity.Role, err error)
	GetByName(ctx context.Context, name string) (role *entity.Role, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error)
	Update(ctx context.Context, role entity.Role) (err error)
	Delete(ctx context.Context, id int) (err error)
	DeleteRoleHasPermissions(ctx context.Context, id int) (err error)
}

type RoleRepositoryImpl struct {
	roleTableName string
	db            *gorm.DB
}

var (
	roleTableName          string = "roles"
	roleRepositoryImpl     *RoleRepositoryImpl
	roleRepositoryImplOnce sync.Once
)

func NewRoleRepository() RoleRepository {
	roleRepositoryImplOnce.Do(func() {
		roleRepositoryImpl = &RoleRepositoryImpl{
			roleTableName: roleTableName,
			db:            connector.LoadDatabase(),
		}
	})
	return roleRepositoryImpl
}

func (repo RoleRepositoryImpl) Create(ctx context.Context, role entity.Role, permissionsID []int) (id int, err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&role)
		if res.Error != nil {
			tx.Rollback()
			return res.Error
		}
		id = role.ID

		for _, permissionID := range permissionsID {
			roleHasPermissionEntity := entity.RoleHasPermission{
				RoleID:       id,
				PermissionID: permissionID,
			}
			if err := tx.Create(&roleHasPermissionEntity).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo RoleRepositoryImpl) GetByID(ctx context.Context, id int) (role *entity.Role, err error) {
	role = &entity.Role{}
	result := repo.db.Where("id = ?", id).First(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role, nil
}

func (repo RoleRepositoryImpl) GetByName(ctx context.Context, name string) (role *entity.Role, err error) {
	role = &entity.Role{}
	result := repo.db.Where("name = ?", name).First(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return role, nil
}

func (repo RoleRepositoryImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error) {
	var count int64
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

func (repo RoleRepositoryImpl) Update(ctx context.Context, role entity.Role) (err error) {
	err = repo.db.Transaction(func(tx *gorm.DB) error {
		var oldData entity.Permission
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

func (repo RoleRepositoryImpl) Delete(ctx context.Context, id int) (err error) {
	deleted := entity.Role{}
	result := repo.db.Where("id = ?", id).Delete(&deleted)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (repo RoleRepositoryImpl) DeleteRoleHasPermissions(ctx context.Context, id int) (err error) {
	deleted := entity.RoleHasPermission{}
	result := repo.db.Where("role_id = ?", id).Delete(&deleted)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

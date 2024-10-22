package dao

import (
	"context"
	"distributed-object-storage/pkg/db/dbm"
	"gorm.io/gorm"
)

type User struct {
	*Base
	ctx  context.Context
	user *dbm.UserInfo
}

func NewUser(db *gorm.DB) *User {
	return &User{
		Base: &Base{DB: db},
		user: &dbm.UserInfo{},
	}
}

// FindAllUser 得到所有的用户列表
func (obj *User) FindAllUser() (results []*dbm.UserInfo, err error) {
	results = []*dbm.UserInfo{}
	if err := obj.DB.Model(&dbm.UserInfo{}).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, err
}

func (obj *User) GetUserInfoByName(UserName string) (tmp *dbm.UserInfo, err error) {
	if err := obj.DB.Model(&dbm.UserInfo{}).Where("username = ?", UserName).First(&tmp).Error; err != nil {
		return nil, err
	}
	return tmp, err
}

func (obj *User) GetUserInfoByID(id uint) (tmp *dbm.UserInfo, err error) {
	err = obj.DB.Model(&dbm.UserInfo{}).Where("id = ?", id).First(&tmp).Error
	if err != nil {
		return nil, err
	}
	return tmp, nil
}

func (obj *User) CreateUser(ctx context.Context, user *dbm.UserInfo) (err error) {
	return obj.DB.Model(&dbm.UserInfo{}).WithContext(ctx).Create(&user).Error
}

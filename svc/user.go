package svc

import (
	"context"
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/pkg/db/dbm"
	"distributed-object-storage/types"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type UserSvc struct {
	nameCache map[int64]string
	lock      sync.RWMutex
	userDao   *dao.User
}

func NewUserSvc(s *dao.S) *UserSvc {
	return &UserSvc{
		nameCache: make(map[int64]string),
		lock:      sync.RWMutex{},
		userDao:   s.User,
	}
}

// FindAllUser 得到所有的用户列表
func (UserSvc *UserSvc) FindAllUser() ([]*dbm.UserInfo, error) {
	userList, err := UserSvc.userDao.FindAllUser()
	if err != nil {
		return nil, err
	}
	return userList, err
}

// GetUserInfoByName 根据用户名得到用户的相关信息
func (UserSvc *UserSvc) GetUserInfoByName(UserName string) (*dbm.UserInfo, error) {
	userinfo, err := UserSvc.userDao.GetUserInfoByName(UserName)
	if err != nil {
		return nil, err
	}
	return userinfo, err
}

// GetUserInfoByID 根据用户名id得到用户的相关信息
func (UserSvc *UserSvc) GetUserInfoByID(UserID uint) (*dbm.UserInfo, error) {
	userinfo, err := UserSvc.userDao.GetUserInfoByID(UserID)
	if err != nil {
		return nil, err
	}
	return userinfo, err
}

func (UserSvc *UserSvc) CreateUser(ctx context.Context, userInfo *types.UserMetaData) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = UserSvc.userDao.CreateUser(ctx, &dbm.UserInfo{
		UserName: userInfo.UserName,
		PassWord: string(hashedPassword),
	})
	if err != nil {
		return err
	}
	return nil
}

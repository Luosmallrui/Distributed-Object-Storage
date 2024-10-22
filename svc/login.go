package svc

import (
	"distributed-object-storage/pkg/db/dao"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"hash/fnv"
	"sync"
)

type AuthSvc struct {
	nameCache map[int64]string
	lock      sync.RWMutex
}

func NewAuthSvc(s *dao.S) *AuthSvc {
	return &AuthSvc{
		nameCache: make(map[int64]string),
		lock:      sync.RWMutex{},
	}
}

// AuthenticateUser 判断登陆输入的账号密码是否正确
func (l *AuthSvc) AuthenticateUser(inputPassword string, dbPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(inputPassword))
	return err == nil
}

// hashPassword 使用FNV哈希算法对密码进行哈希处理，并返回哈希值的字符串表示形式
func hashPassword(password string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(password))
	hashed := h.Sum64()
	return fmt.Sprintf("%x", hashed)
}

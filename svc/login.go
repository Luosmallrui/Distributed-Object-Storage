package svc

import (
	"distributed-object-storage/pkg/db/dao"
	"fmt"
	"hash/fnv"
	"sync"
)

type LoginSvc struct {
	nameCache map[int64]string
	lock      sync.RWMutex
}

func NewLoginSvc(s *dao.S) *LoginSvc {
	return &LoginSvc{
		nameCache: make(map[int64]string),
		lock:      sync.RWMutex{},
	}
}

// AuthenticateUser 判断登陆输入的账号密码是否正确
func (loginSvc *LoginSvc) AuthenticateUser(password string, entryPassword string) bool {
	isValid := verifyPassword(entryPassword, password)
	return isValid
}

// hashPassword 使用FNV哈希算法对密码进行哈希处理，并返回哈希值的字符串表示形式
func hashPassword(password string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(password))
	hashed := h.Sum64()
	return fmt.Sprintf("%x", hashed)
}

// verifyPassword 用于验证加密后的密码是否与原始密码匹配
func verifyPassword(encryptedPassword, originalPassword string) bool {
	return encryptedPassword == hashPassword(originalPassword)
}

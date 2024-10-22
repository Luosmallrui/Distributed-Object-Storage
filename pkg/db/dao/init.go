package dao

import (
	"distributed-object-storage/pkg/db"
	"gorm.io/gorm"
)

type S struct {
	*Base
	DB           *gorm.DB
	MetadataNode MetadataNode
}

func Init() *S {
	return &S{
		Base: &Base{
			//DB: db.Db(),
		},
		DB:           db.Db(),
		MetadataNode: MetadataNode{},
	}
}

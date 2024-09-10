package dao

import (
	"context"
	"distributed-object-storage/pkg/db/dbm"
	"gorm.io/gorm"
)

type MetadataNode struct {
	*Base
	ctx           context.Context
	emergencyPlan *dbm.ObjectMetadata
}

func NewMetadataNode(db *gorm.DB) *MetadataNode {
	return &MetadataNode{
		Base:          &Base{DB: db},
		emergencyPlan: &dbm.ObjectMetadata{},
	}
}

func (obj *MetadataNode) FindAllEmergencyPlan() (results []*dbm.ObjectMetadata, err error) {
	results = []*dbm.ObjectMetadata{}
	if err := obj.DB.Model(&dbm.ObjectMetadata{}).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, err
}

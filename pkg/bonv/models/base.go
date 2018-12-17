package models

import (
	"yunion.io/x/pkg/util/stringutils"

	"yunion.io/x/onecloud/pkg/cloudcommon/db"
)

type SResourceBase struct {
	db.SResourceBase

	Id string `width:"128" charset:"ascii" primary:"true" list:"user"`
}

func (model *SResourceBase) BeforeInsert() {
	if len(model.Id) == 0 {
		model.Id = stringutils.UUID4()
	}
}

type SResourceBaseManager struct {
	db.SResourceBaseManager
}

func NewResourceBaseManager(dt interface{}, tableName string, keyword string, keywordPlural string) SResourceBaseManager {
	return SResourceBaseManager{
		db.NewResourceBaseManager(dt, tableName, keyword, keywordPlural),
	}
}

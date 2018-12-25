package models

import (
	"yunion.io/x/pkg/util/stringutils"
	"yunion.io/x/sqlchemy"

	"yunion.io/x/onecloud/pkg/bonv/cloud"
	"yunion.io/x/onecloud/pkg/bonv/cloud/types"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
)

type SResourceBase struct {
	db.SResourceBase

	Id string `width:"128" charset:"ascii" primary:"true" list:"user"`
}

func (r *SResourceBase) BeforeInsert() {
	if len(r.Id) == 0 {
		r.Id = stringutils.UUID4()
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

func (man *SResourceBaseManager) FetchById(id string) (db.IModel, error) {
	return db.FetchById(man, id)
}

func (man *SResourceBaseManager) FilterById(q *sqlchemy.SQuery, id string) *sqlchemy.SQuery {
	return q.Equals("id", id)
}

func (man *SResourceBaseManager) FilterByNotId(q *sqlchemy.SQuery, id string) *sqlchemy.SQuery {
	return q.NotEquals("id", id)
}

func (man *SResourceBaseManager) FilterByName(q *sqlchemy.SQuery, name string) *sqlchemy.SQuery {
	return q.Equals("name", name)
}

type SCloudResourceInfraMixin struct {
	IsInfra bool ` primary:"true" list:"user"`
}

type SResourceAccountMixin struct {
	AccountId string `width:"36" charset:"ascii" nullable:"false"`
}

func (r *SResourceAccountMixin) getClient() (types.Client, error) {
	q := CloudAccountManager.Query().Equals("id", r.AccountId)
	account := &SCloudAccount{}
	err := q.First(account)
	if err != nil {
		return nil, err
	}
	client, err := cloud.NewClient(account.Provider, account.Account, account.Secret)
	return client, err
}

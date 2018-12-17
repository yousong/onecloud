package models

import (
	"context"
	"database/sql"

	"yunion.io/x/onecloud/pkg/bonv/types"
)

type SCloudAccount struct {
	SResourceBase

	Provider string `width:"36" charset:"ascii" nullable:"false"`
	Account  string `width:"36" charset:"ascii" nullable:"false"`
	Secret   string `width:"36" charset:"ascii" nullable:"false"`
}

type SCloudAccountManager struct {
	SResourceBaseManager
}

var CloudAccountManager *SCloudAccountManager

func init() {
	CloudAccountManager = &SCloudAccountManager{
		SResourceBaseManager: NewResourceBaseManager(
			SCloudAccount{},
			"cloud_accounts_tbl",
			"cloud_account",
			"cloud_accounts",
		),
	}
}

func (man *SCloudAccountManager) UpdateOrNewFromRequest(ctx context.Context, accountInfo *types.SCloudConnectRequestVpcAccount) (*SCloudAccount, error) {
	{
		// fetch
		account := &SCloudAccount{}
		q := man.Query().
			Equals("provider", accountInfo.Provider).
			Equals("account", accountInfo.Account)
		err := q.First(account)
		if err == nil {
			// update secret
			if account.Secret != accountInfo.Secret {
				_, err := CloudAccountManager.TableSpec().Update(account, func() error {
					account.Secret = accountInfo.Secret
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
			return account, nil
		} else {
			if err != sql.ErrNoRows {
				return nil, err
			}
		}
		// not found
	}
	{
		// insert new
		account := &SCloudAccount{
			Provider: accountInfo.Provider,
			Account:  accountInfo.Account,
			Secret:   accountInfo.Secret,
		}
		account.SetModelManager(CloudAccountManager)
		err := CloudAccountManager.TableSpec().Insert(account)
		if err != nil {
			return nil, err
		}
		return account, nil
	}
}

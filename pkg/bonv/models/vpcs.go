package models

import (
	"context"
	"database/sql"

	"yunion.io/x/onecloud/pkg/bonv/utils"
)

type SVpc struct {
	SResourceBase

	ExternalId  string `width:"36" charset:"ascii" nullable:"false"`
	Name        string `width:"36" charset:"ascii" nullable:"false"`
	Description string

	CidrBlock string
	Status    string
}

type SVpcManager struct {
	SResourceBaseManager
}

var VpcManager *SVpcManager

func init() {
	VpcManager = &SVpcManager{
		SResourceBaseManager: NewResourceBaseManager(
			SVpc{},
			"vpcs_tbl",
			"vpc",
			"vpcs",
		),
	}
}

func (man *SVpcManager) UpdateOrNewFromCloud(ctx context.Context, cloudVpc *utils.Vpc) (*SVpc, error) {
	{
		// fetch
		vpc := &SVpc{}
		q := man.Query().Equals("external_id", cloudVpc.Id)
		err := q.First(vpc)
		if err == nil {
			// update secret
			_, err := VpcManager.TableSpec().Update(vpc, func() error {
				vpc.ExternalId = cloudVpc.Id
				vpc.Name = cloudVpc.Name
				vpc.Description = cloudVpc.Description
				vpc.CidrBlock = cloudVpc.CidrBlock
				vpc.Status = cloudVpc.Status
				return nil
			})
			if err != nil {
				return nil, err
			}
			return vpc, nil
		} else {
			if err != sql.ErrNoRows {
				return nil, err
			}
		}
		// not found
	}
	{
		// insert new
		vpc := &SVpc{
			ExternalId:  cloudVpc.Id,
			Name:        cloudVpc.Name,
			Description: cloudVpc.Description,

			CidrBlock: cloudVpc.CidrBlock,
			Status:    cloudVpc.Status,
		}
		vpc.SetModelManager(VpcManager)
		err := VpcManager.TableSpec().Insert(vpc)
		if err != nil {
			return nil, err
		}
		return vpc, nil
	}
}

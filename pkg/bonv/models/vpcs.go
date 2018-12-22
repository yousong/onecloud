package models

import (
	"context"
	"database/sql"
	"fmt"

	"yunion.io/x/onecloud/pkg/bonv/utils"
)

type SVpc struct {
	SResourceBase

	Provider    string `width:"36" charset:"ascii" nullable:"false"`
	RegionId    string `width:"36" charset:"ascii" nullable:"false"`
	ExternalId  string `width:"36" charset:"ascii" nullable:"false"`
	Name        string `width:"36" charset:"ascii" nullable:"false"`
	Description string

	CidrBlock string `width:"36" charset:"ascii" nullable:"false"`
	VRouterId string `width:"36" charset:"ascii" nullable:"false"`
	Status    string `width:"36" charset:"ascii" nullable:"false"`

	AccountId string `width:"36" charset:"ascii" nullable:"false"`
	IsInfra   bool   `charset:"ascii" nullable:"false"`
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
				vpc.Provider = cloudVpc.Provider
				vpc.RegionId = cloudVpc.RegionId
				vpc.ExternalId = cloudVpc.Id
				vpc.Name = cloudVpc.Name
				vpc.Description = cloudVpc.Description
				vpc.CidrBlock = cloudVpc.CidrBlock
				vpc.VRouterId = cloudVpc.VRouterId
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
			Provider:    cloudVpc.Provider,
			RegionId:    cloudVpc.RegionId,
			ExternalId:  cloudVpc.Id,
			Name:        cloudVpc.Name,
			Description: cloudVpc.Description,

			CidrBlock: cloudVpc.CidrBlock,
			VRouterId: cloudVpc.VRouterId,
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

func (vpc *SVpc) UniqId() string {
	return fmt.Sprintf("%s:vpc:%s", vpc.Provider, vpc.ExternalId)
}

func (vpc *SVpc) connectInfra(ctx context.Context) error {
	// TODO move this check to outer
	if vpc.IsInfra {
		return fmt.Errorf("infra vpc is not supposed to be initiating connect")
	}
	q := VpcManager.Query().
		Equals("provider", vpc.Provider).
		IsFalse("is_infra")
	infraVpcs := []SVpc{}
	// TODO fetch models
	ok := false
	for i := range infraVpcs {
		infraVpc := &infraVpcs[i]
		err := VpcConnectManager.createVpcConnect(ctx, vpc, infraVpc)
		if err != nil {
			continue
		}
		ok = true
		break
	}
	if !ok {
		// call admin!
		return fmt.Errorf("tried %d our infra vpc, all unavailable", len(infraVpcs))
	}
	return nil
}

func (vpc *SVpc) toCloudVpc() *utils.Vpc {
	cloudVpc := &utils.Vpc{
		Provider:    vpc.Provider,
		RegionId:    vpc.RegionId,
		Id:          vpc.ExternalId,
		Name:        vpc.Name,
		Description: vpc.Description,

		CidrBlock: vpc.CidrBlock,
		VRouterId: vpc.VRouterId,
		Status:    vpc.Status,
	}
	return cloudVpc
}

package models

import (
	"context"
	"database/sql"
	"fmt"

	"yunion.io/x/log"
	"yunion.io/x/sqlchemy"

	"yunion.io/x/onecloud/pkg/bonv/cloud/types"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
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

	SCloudResourceInfraMixin
	SResourceAccountMixin
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

func (man *SVpcManager) UpdateOrNewFromCloud(ctx context.Context, cloudVpc *types.Vpc, account *SCloudAccount) (*SVpc, error) {
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
				vpc.AccountId = account.Id
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
			SResourceAccountMixin: SResourceAccountMixin{
				AccountId: account.Id,
			},
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

// 允许幂等操作，若已连接，不做操作，返回连接成功
func (vpc *SVpc) connectInfra(ctx context.Context) error {
	// TODO move this check to outer
	if vpc.IsInfra {
		return fmt.Errorf("infra vpc is not supposed to be initiating connect")
	}
	if maybe, vc, err := vpc.maybeConnected(ctx); maybe {
		if vc != nil {
			log.Warningf("vpc %s is already connected: %s", vpc.UniqId(), vc.UniqId())
			return nil
		} else {
			return err
		}
	}
	return vpc.connectInfra_(ctx)
}

func (vpc *SVpc) connectInfra_(ctx context.Context) error {
	q := VpcManager.Query().
		Equals("provider", vpc.Provider).
		Equals("region_id", vpc.RegionId).
		IsTrue("is_infra")
	infraVpcs := []SVpc{}
	if err := db.FetchModelObjects(VpcManager, q, &infraVpcs); err != nil {
		return fmt.Errorf("querying candidate peer vpcs: %s", err)
	}
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
		// call site admin!
		return fmt.Errorf("tried %d of our infra vpcs, all unavailable", len(infraVpcs))
	}
	return nil
}

func (vpc *SVpc) maybeConnected(ctx context.Context) (bool, *SVpcConnect, error) {
	q := VpcConnectManager.Query()
	q = q.Equals("provider", vpc.Provider).
		Filter(sqlchemy.OR(
			sqlchemy.Equals(q.Field("vpc_id0"), vpc.Id),
			sqlchemy.Equals(q.Field("vpc_id1"), vpc.Id),
		))
	vc := &SVpcConnect{}
	vc.SetModelManager(VpcConnectManager)
	err := q.First(vc)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil
		}
		return true, nil, fmt.Errorf("maybe connected: %s", err)
	}
	return true, vc, nil
}

func (vpc *SVpc) toCloudVpc() *types.Vpc {
	cloudVpc := &types.Vpc{
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

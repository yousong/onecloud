package models

import (
	"context"
	"database/sql"
	"fmt"

	"yunion.io/x/onecloud/pkg/bonv/utils"
)

type SVpcConnect struct {
	SResourceBase

	Provider string `width:"36" charset:"ascii" nullable:"false"`
	VpcId0   string `width:"36" charset:"ascii" nullable:"false"`
	VpcId1   string `width:"36" charset:"ascii" nullable:"false"`
}

type SVpcConnectManager struct {
	SResourceBaseManager
}

var VpcConnectManager *SVpcConnectManager

func init() {
	VpcConnectManager = &SVpcConnectManager{
		SResourceBaseManager: NewResourceBaseManager(
			SVpcConnect{},
			"vpc_connects_tbl",
			"vpc_connect",
			"vpc_connects",
		),
	}
}

func (vpc *SVpcConnect) createVpcConnect(ctx context.Context, vpc, infraVpc *SVpc) error {
	q := VpcManager.Query().
		Equals("vpc_id0", vpc.Id).
		Equals("vpc_id1", infraVpc.Id)
	if q.Count() > 0 {
		return fmt.Errorf("vpc connect between %s and %s already exist", vpc.UniqId(), infraVpc.UniqId())
	}
	cloudVpc := vpc.toCloudVpc()
	cloudInfraVpc := infraVpc.toCloudVpc()
	client := vpc.getClient()
	return nil
}

package models

import (
	"context"
	"fmt"

	"yunion.io/x/onecloud/pkg/bonv/cloud/types"
)

type SVpcConnect struct {
	SResourceBase

	Provider           string `width:"36" charset:"ascii" nullable:"false"`
	VpcId0             string `width:"36" charset:"ascii" nullable:"false"`
	VpcId1             string `width:"36" charset:"ascii" nullable:"false"`
	RouterInterfaceId0 string `width:"36" charset:"ascii" nullable:"false"`
	RouterInterfaceId1 string `width:"36" charset:"ascii" nullable:"false"`
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

func (man *SVpcConnectManager) createVpcConnect(ctx context.Context, vpc, infraVpc *SVpc) error {
	if vpc.Provider != infraVpc.Provider {
		return fmt.Errorf("vpc connect can only happen with in the same provider, got %s and %s", vpc.UniqId(), infraVpc.UniqId())
	}
	q := VpcManager.Query().
		Equals("vpc_id0", vpc.Id).
		Equals("vpc_id1", infraVpc.Id)
	if q.Count() > 0 {
		return fmt.Errorf("vpc connect between %s and %s already exist", vpc.UniqId(), infraVpc.UniqId())
	}
	cloudVpc := vpc.toCloudVpc()
	cloudInfraVpc := infraVpc.toCloudVpc()
	client, err := vpc.getClient()
	if err != nil {
		return err
	}
	infraClient, err := infraVpc.getClient()
	if err != nil {
		return err
	}
	{
		resp, err := client.ConnectVpc(&types.ConnectVpcRequest{
			Vpc:             cloudVpc,
			AcceptingVpc:    cloudInfraVpc,
			AcceptingClient: infraClient,
		})
		if err != nil {
			return err
		}
		vc := &SVpcConnect{
			VpcId0:             vpc.Id,
			RouterInterfaceId0: resp.InitiatingRouterInterfaceId,

			VpcId1:             infraVpc.Id,
			RouterInterfaceId1: resp.AcceptingRouterInterfaceId,
		}
		vc.SetModelManager(VpcConnectManager)
		if err := VpcConnectManager.TableSpec().Insert(vc); err != nil {
			return fmt.Errorf("insert vpc connect: %s", err)
		}
	}
	return nil
}

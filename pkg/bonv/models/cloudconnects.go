package models

import (
	"context"
	"fmt"
	"net/http"

	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/appsrv"
	"yunion.io/x/onecloud/pkg/bonv/types"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
)

func HandleNewCloudConnectRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := &types.SCloudConnectRequest{}
	if err := appsrv.FetchStruct(r, req); err != nil {
		httperrors.InvalidInputError(w, "parsing request: %s", err)
		return
	}
	if err := req.Validate(); err != nil {
		httperrors.InvalidInputError(w, "validate: %s", err)
		return
	}
	if err := CloudConnectManager.CreateFromRequest(ctx, req); err != nil {
		httperrors.BadRequestError(w, "%s", err)
		return
	}
	w.Write([]byte("未完待续\n"))
}

type SCloudConnectManager struct {
	SResourceBaseManager
}

var CloudConnectManager *SCloudConnectManager

func init() {
	CloudConnectManager = &SCloudConnectManager{
		SResourceBaseManager: NewResourceBaseManager(
			SCloudConnect{},
			"cloud_connects_tbl",
			"cloud_connect",
			"cloud_connects",
		),
	}
}

type SCloudConnect struct {
	SResourceBase

	Provider0 string `width:"36" charset:"ascii" nullable:"false"`
	Provider1 string `width:"36" charset:"ascii" nullable:"false"`

	AccountId0 string `width:"36" charset:"ascii" nullable:"false"`
	AccountId1 string `width:"36" charset:"ascii" nullable:"false"`

	VpcId0 string `width:"36" charset:"ascii" nullable:"false"`
	VpcId1 string `width:"36" charset:"ascii" nullable:"false"`

	Phase string `width:"36" charset:"ascii" nullable:"false" default:"init"`
}

func (man *SCloudConnectManager) AllowCreateItem(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return false
}

func (man *SCloudConnectManager) CreateFromRequest(ctx context.Context, req *types.SCloudConnectRequest) error {
	cloudVpcs, err := req.GetVpcs(ctx)
	if err != nil {
		return fmt.Errorf("getting vpc: %s", err)
	}
	accounts := []*SCloudAccount{}
	vpcs := []*SVpc{}
	// now account info is also validated
	for i := range req.Vpcs {
		vpcInfo := &req.Vpcs[i]
		{
			accountInfo := &vpcInfo.Account
			account, err := CloudAccountManager.UpdateOrNewFromRequest(ctx, accountInfo)
			if err != nil {
				return fmt.Errorf("updating account index %d: %s", i, err)
			}
			accounts = append(accounts, account)
		}
		{
			cloudVpc := cloudVpcs[i]
			vpc, err := VpcManager.UpdateOrNewFromCloud(ctx, cloudVpc, accounts[i])
			if err != nil {
				return fmt.Errorf("updating vpc index %d: %s", i, err)
			}
			vpcs = append(vpcs, vpc)
		}
	}
	{
		q := man.Query().
			Equals("vpc_id0", vpcs[0].Id).
			Equals("vpc_id1", vpcs[1].Id)
		if q.Count() > 0 {
			return fmt.Errorf("a connect request between %s and %s already exist", vpcs[0].UniqId(), vpcs[1].UniqId())
		}

		vpcInfo0 := &req.Vpcs[0]
		vpcInfo1 := &req.Vpcs[1]
		connectRequest := &SCloudConnect{
			Provider0:  vpcInfo0.Account.Provider,
			Provider1:  vpcInfo1.Account.Provider,
			AccountId0: accounts[0].Id,
			AccountId1: accounts[1].Id,
			VpcId0:     vpcs[0].Id,
			VpcId1:     vpcs[1].Id,
		}
		connectRequest.SetModelManager(man)
		err := man.TableSpec().Insert(connectRequest)
		if err != nil {
			return fmt.Errorf("inserting cloud connect request: %s", err)
		}
	}
	return nil
}

func (cn *SCloudConnect) AllowPerformConnectInfra(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return true
}

func (cn *SCloudConnect) PerformConnectInfra(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {
	return nil, cn.connectInfra(ctx)
}

func (cn *SCloudConnect) connectInfra(ctx context.Context) error {
	doConnect := func(vpcId string) error {
		m, err := VpcManager.FetchById(vpcId)
		if err != nil {
			return fmt.Errorf("cannot find vpc %s", vpcId)
		}
		vpc := m.(*SVpc)
		if err := vpc.connectInfra(ctx); err != nil {
			return err
		}
		return nil
	}
	if err := doConnect(cn.VpcId0); err != nil {
		return err
	}
	if err := doConnect(cn.VpcId1); err != nil {
		return err
	}
	// TODO, return two vpc connects
	return nil
}

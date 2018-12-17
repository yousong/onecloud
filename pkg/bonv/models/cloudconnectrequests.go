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
	if err := CloudConnectRequestManager.CreateFromRequest(ctx, req); err != nil {
		httperrors.GeneralServerError(w, err)
		return
	}
	w.Write([]byte("未完待续\n"))
}

type SCloudConnectRequestManager struct {
	SResourceBaseManager
}

var CloudConnectRequestManager *SCloudConnectRequestManager

func init() {
	CloudConnectRequestManager = &SCloudConnectRequestManager{
		SResourceBaseManager: NewResourceBaseManager(
			SCloudConnectRequest{},
			"cloud_connect_requests_tbl",
			"cloud_connect_request",
			"cloud_connect_requests",
		),
	}
}

type SCloudConnectRequest struct {
	SResourceBase

	Provider0 string `width:"36" charset:"ascii" nullable:"false"`
	Provider1 string `width:"36" charset:"ascii" nullable:"false"`

	AccountId0 string `width:"36" charset:"ascii" nullable:"false"`
	AccountId1 string `width:"36" charset:"ascii" nullable:"false"`

	VpcId0 string `width:"36" charset:"ascii" nullable:"false"`
	VpcId1 string `width:"36" charset:"ascii" nullable:"false"`
}

func (man *SCloudConnectRequestManager) AllowCreateItem(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return false
}

func (man *SCloudConnectRequestManager) CreateFromRequest(ctx context.Context, req *types.SCloudConnectRequest) error {
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
			vpc, err := VpcManager.UpdateOrNewFromCloud(ctx, cloudVpc)
			if err != nil {
				return fmt.Errorf("updating vpc index %d: %s", i, err)
			}
			vpcs = append(vpcs, vpc)
		}
	}
	{
		vpcInfo0 := &req.Vpcs[0]
		vpcInfo1 := &req.Vpcs[1]
		connectRequest := &SCloudConnectRequest{
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

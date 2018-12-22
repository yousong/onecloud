package types

import (
	"context"
	"fmt"

	"yunion.io/x/onecloud/pkg/bonv/cloud/aliyun"
	"yunion.io/x/onecloud/pkg/bonv/cloud/qcloud"
	"yunion.io/x/onecloud/pkg/bonv/cloud/types"
)

const (
	CLOUD_CONNECT_REQUEST_V1 = "v1"
)

// 用户请求参数
type SCloudConnectRequest struct {
	Version string
	Vpcs    []SCloudConnectRequestVpc
}

func (req *SCloudConnectRequest) Validate() error {
	if req.Version != CLOUD_CONNECT_REQUEST_V1 {
		return fmt.Errorf("expecting request version %s", CLOUD_CONNECT_REQUEST_V1)
	}
	if len(req.Vpcs) != 2 {
		return fmt.Errorf("expecting request to contain 2 vpcs, got %d", len(req.Vpcs))
	}
	{
		m := map[string]int{}
		for i := range req.Vpcs {
			vpc := &req.Vpcs[i]
			provider := vpc.Account.Provider
			if j, ok := m[provider]; ok {
				return fmt.Errorf("vpc index %d and index %d are from the same provider %s", j, i, provider)
			}
			m[provider] = i
		}
	}
	return nil
}

// 也用于验证参数中账号和Vpc信息是否有效
func (req *SCloudConnectRequest) GetVpcs(ctx context.Context) ([]*types.Vpc, error) {
	vpcs := []*types.Vpc{}
	for i := range req.Vpcs {
		reqVpc := &req.Vpcs[i]
		vpc, err := reqVpc.GetVpc()
		if err != nil {
			return nil, fmt.Errorf("failed getting vpc of index %d: %s", i, err)
		}
		vpcs = append(vpcs, vpc)
	}
	return vpcs, nil
}

type SCloudConnectRequestVpc struct {
	Account SCloudConnectRequestVpcAccount
	Vpc     SCloudConnectRequestVpcParams
}

func (reqVpc *SCloudConnectRequestVpc) GetClient() (client types.Client, err error) {
	account := &reqVpc.Account
	switch account.Provider {
	case types.PROVIDER_ALIYUN:
		client, err = aliyun.NewClient(account.Account, account.Secret)
		if err != nil {
			return nil, fmt.Errorf("failed creating aliyun client: %s", err)
		}
	case types.PROVIDER_TENCENTCLOUD:
		client, err = qcloud.NewClient(account.Account, account.Secret)
		if err != nil {
			return nil, fmt.Errorf("failed creating tencentcloud client: %s", err)
		}
	default:
		return nil, fmt.Errorf("account provider not supported yet: %q", account.Provider)
	}
	return client, nil
}

func (reqVpc *SCloudConnectRequestVpc) GetVpc() (vpc *types.Vpc, err error) {
	client, err := reqVpc.GetClient()
	if err != nil {
		return nil, err
	}
	vpcParams := &reqVpc.Vpc
	req := &types.DescribeVpcRequest{
		RegionId: vpcParams.RegionId,
		VpcId:    vpcParams.VpcId,
	}
	//fmt.Printf("req params %#v\n", req)
	vpc, err = client.DescribeVpc(req)
	return vpc, err
}

// TODO STS token
type SCloudConnectRequestVpcAccount struct {
	Provider string
	Account  string
	Secret   string
}

type SCloudConnectRequestVpcParams struct {
	RegionId  string
	VpcId     string
	VSwitchId string
	Cidrs     []string
}

package qcloud

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"yunion.io/x/onecloud/pkg/bonv/cloud/types"
)

type Client struct {
	credential *common.Credential
}

func NewClient(accessKeyId, accessKeySecret string) (types.Client, error) {
	client := &Client{
		credential: common.NewCredential(accessKeyId, accessKeySecret),
	}
	return client, nil
}

func (client *Client) cvmClient(regionId string) (*cvm.Client, error) {
	cpf := profile.NewClientProfile()
	cli, err := cvm.NewClient(client.credential, regions.Guangzhou, cpf)
	return cli, err
}

func (client *Client) vpcClient(regionId string) (*vpc.Client, error) {
	cpf := profile.NewClientProfile()
	cli, err := vpc.NewClient(client.credential, regionId, cpf)
	return cli, err
}

func (client *Client) DoUsuableTest() (bool, error) {
	cli, err := client.cvmClient(regions.Beijing)
	if err != nil {
		return false, fmt.Errorf("making cvm client: %s", err)
	}
	req := cvm.NewDescribeRegionsRequest()
	_, err = cli.DescribeRegions(req)
	if err != nil {
		return false, fmt.Errorf("DescribeRegions: %s", err)
	}
	return true, nil
}

func (client *Client) DescribeVpc(r *types.DescribeVpcRequest) (*types.Vpc, error) {
	cli, err := client.vpcClient(r.RegionId)
	if err != nil {
		return nil, fmt.Errorf("making vpc client: %s", err)
	}
	req := vpc.NewDescribeVpcsRequest()
	req.VpcIds = []*string{&r.VpcId}
	resp, err := cli.DescribeVpcs(req)
	if err != nil {
		return nil, fmt.Errorf("DescribeVpcs: %s", err)
	}
	respR := resp.Response
	if len(respR.VpcSet) != 1 {
		return nil, fmt.Errorf("expecting 1 vpc in response, got %d", len(respR.VpcSet))
	}
	vpcR := respR.VpcSet[0]
	vpc := &types.Vpc{
		Id:          StringV(vpcR.VpcId),
		Name:        StringV(vpcR.VpcName),
		Description: "",

		Provider:   types.PROVIDER_TENCENTCLOUD,
		RegionId:   r.RegionId,
		Status:     "",
		CidrBlock:  StringV(vpcR.CidrBlock),
		VRouterId:  "",
		VSwitchIds: nil,
	}
	if err := vpc.Validate(); err != nil {
		return nil, err
	}
	return vpc, nil
}

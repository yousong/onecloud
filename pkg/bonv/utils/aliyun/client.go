package aliyun

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"yunion.io/x/onecloud/pkg/bonv/utils"
)

type Client struct {
	*sdk.Client
}

//type ClientConfig sdk.Config

func NewClient(accessKeyId, accessKeySecret string) (utils.Client, error) {
	regionId := "cn-beijing"
	config := sdk.NewConfig()
	credential := &credentials.BaseCredential{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
	client, err := sdk.NewClientWithOptions(regionId, config, credential)
	if err != nil {
		return nil, err
	}
	return &Client{Client: client}, nil
}

func (client *Client) DoUsuableTest() (bool, error) {
	req := ecs.CreateDescribeRegionsRequest()
	resp := ecs.CreateDescribeRegionsResponse()
	err := client.DoAction(req, resp)
	if err != nil {
		return false, fmt.Errorf("DescribeRegions: %s", err)
	}
	return true, nil
}

func (client *Client) DescribeVpc(r *utils.DescribeVpcRequest) (*utils.Vpc, error) {
	req := vpc.CreateDescribeVpcAttributeRequest()
	req.RpcRequest.RegionId = r.RegionId
	req.VpcId = r.VpcId
	resp := vpc.CreateDescribeVpcAttributeResponse()
	err := client.DoAction(req, resp)
	if err != nil {
		return nil, err
	}
	vpc := &utils.Vpc{
		Id:          resp.VpcId,
		Name:        resp.VpcName,
		Description: resp.Description,
		Status:      resp.Status,

		CidrBlock:  resp.CidrBlock,
		VSwitchIds: resp.VSwitchIds.VSwitchId,
	}
	if err := vpc.Validate(); err != nil {
		return nil, err
	}
	return vpc, nil
}

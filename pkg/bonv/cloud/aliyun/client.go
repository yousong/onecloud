package aliyun

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"yunion.io/x/onecloud/pkg/bonv/cloud/types"
)

type Client struct {
	credential auth.Credential
}

//type ClientConfig sdk.Config

func NewClient(accessKeyId, accessKeySecret string) (types.Client, error) {
	client := &Client{
		credential: &credentials.BaseCredential{
			AccessKeyId:     accessKeyId,
			AccessKeySecret: accessKeySecret,
		},
	}
	return client, nil
}

func (client *Client) ecsClient() (*ecs.Client, error) {
	config := sdk.NewConfig()
	srvClient, err := ecs.NewClientWithOptions("cn-beijing", config, client.credential)
	if err != nil {
		return nil, fmt.Errorf("create ecs client: %s", err)
	}
	return srvClient, nil
}

func (client *Client) vpcClient() (*vpc.Client, error) {
	config := sdk.NewConfig()
	srvClient, err := vpc.NewClientWithOptions("cn-beijing", config, client.credential)
	if err != nil {
		return nil, fmt.Errorf("create vpc client: %s", err)
	}
	return srvClient, nil
}

func (client *Client) DoUsuableTest() (bool, error) {
	ecsClient, err := client.ecsClient()
	if err != nil {
		return false, err
	}
	req := ecs.CreateDescribeRegionsRequest()
	_, err = ecsClient.DescribeRegions(req)
	if err != nil {
		return false, fmt.Errorf("DescribeRegions: %s", err)
	}
	return true, nil
}

func (client *Client) DescribeVpc(r *types.DescribeVpcRequest) (*types.Vpc, error) {
	req := vpc.CreateDescribeVpcAttributeRequest()
	req.RpcRequest.RegionId = r.RegionId
	req.VpcId = r.VpcId
	vpcClient, err := client.vpcClient()
	if err != nil {
		return nil, err
	}
	resp, err := vpcClient.DescribeVpcAttribute(req)
	if err != nil {
		return nil, err
	}
	vpc := &types.Vpc{
		Id:          resp.VpcId,
		Name:        resp.VpcName,
		Description: resp.Description,

		Provider:   types.PROVIDER_ALIYUN,
		RegionId:   resp.RegionId,
		Status:     resp.Status,
		CidrBlock:  resp.CidrBlock,
		VRouterId:  resp.VRouterId,
		VSwitchIds: resp.VSwitchIds.VSwitchId,
	}
	if err := vpc.Validate(); err != nil {
		return nil, err
	}
	return vpc, nil
}

func (client *Client) createRouterIface(vpcA, vpcB *types.Vpc, role string) (*vpc.CreateRouterInterfaceResponse, error) {
	vpcClient, err := client.vpcClient()
	if err != nil {
		return nil, err
	}
	req := vpc.CreateCreateRouterInterfaceRequest()
	req.RegionId = vpcA.RegionId
	req.RouterType = "VRouter"
	req.RouterId = vpcA.VRouterId
	req.Role = role
	req.Spec = "Xlarge.1"
	req.OppositeRegionId = vpcB.RegionId
	req.OppositeRouterType = "VRouter"
	req.OppositeRouterId = vpcB.VRouterId
	respChan, errChan := vpcClient.CreateRouterInterfaceWithChan(req)
	select {
	case resp := <-respChan:
		return resp, nil
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("initiating side: create router interface: %s", err)
		}
	}
	return nil, fmt.Errorf("unreachable code!")
}

func (client *Client) deleteRouterIface(regionId, routerIfaceId string) (*vpc.DeleteRouterInterfaceResponse, error) {
	vpcClient, err := client.vpcClient()
	if err != nil {
		return nil, err
	}
	req := vpc.CreateDeleteRouterInterfaceRequest()
	req.RegionId = regionId
	req.RouterInterfaceId = routerIfaceId
	resp, err := vpcClient.DeleteRouterInterface(req)
	return resp, err
}

func (client *Client) ConnectVpc(r *types.ConnectVpcRequest) (*types.ConnectVpcResponse, error) {
	initiatingVpc := r.Vpc
	acceptingVpc := r.AcceptingVpc

	initiatingResp, err := client.createRouterIface(initiatingVpc, acceptingVpc, "InitiatingSide")
	if err != nil {
		return nil, err
	}

	acceptingClient := r.AcceptingClient.(*Client)
	acceptingResp, err := acceptingClient.createRouterIface(acceptingVpc, initiatingVpc, "AcceptingSide")
	if err != nil {
		client.deleteRouterIface(initiatingVpc.RegionId, initiatingResp.RouterInterfaceId)
		return nil, err
	}
	resp := &types.ConnectVpcResponse{
		InitiatingRegionId:          initiatingVpc.RegionId,
		InitiatingRouterInterfaceId: initiatingResp.RouterInterfaceId,
		AcceptingRegionId:           acceptingVpc.RegionId,
		AcceptingRouterInterfaceId:  acceptingResp.RouterInterfaceId,
	}
	return resp, nil
}

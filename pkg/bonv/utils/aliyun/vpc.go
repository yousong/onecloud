package aliyun

import (
	"time"
)

type DescribeVpcAttributeRequest struct {
	// 是	要执行的操作。 取值：DescribeVpcAttribute
	Action string

	// 是	要查询的VPC ID。
	VpcId string

	// 是	VPC的所属地域ID。
	RegionId string

	//否	是否是默认VPC，取值：
	//false：不是默认VPC
	//true：是默认VPC
	IsDefault bool
}

type DescribeVpcAttributeResponse struct {
	// 请求ID。
	RequestId string

	// VPC的名称。
	VpcName string

	// VPC的描述。
	Description string

	// 是否是默认VPC。
	IsDefault bool

	// VPC的路由器ID。
	VRouterId string

	// VPC ID。
	VpcId string

	// VPC的私网网段。
	CidrBlock string

	// VPC下的交换机列表。
	VSwitchIds DescribeVpcAttributeResponseVSwitchIds

	// VPC的状态。
	Status string

	// VPC的创建时间。
	CreationTime time.Time

	// VPC下的资源列表。
	//CloudResources List
}

type DescribeVpcAttributeResponseVSwitchIds struct {
	VSwitchId []string
}

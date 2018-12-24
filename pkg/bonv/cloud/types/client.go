package types

const (
	PROVIDER_ALIYUN       = "aliyun"
	PROVIDER_TENCENTCLOUD = "tencentcloud"
)

type Client interface {
	DoUsuableTest() (bool, error)
	DescribeVpc(*DescribeVpcRequest) (*Vpc, error)
	ConnectVpc(*ConnectVpcRequest) (*ConnectVpcResponse, error)
}

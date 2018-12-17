package options

type BonvRequestVpcOptions struct {
	Manager struct {
		Provider string `help:"cloud provider name" choices:"aliyun|qcloud"`
		Account  string `help:"cloud account (access key)"`
		Secret   string `help:"cloud account secret (access secret)"`
	}
	Vpc struct {
		RegionId string   `help:"region id of the vpc"`
		VpcId    string   `help:"vpc id"`
		Cidrs    []string `help:"cidrs to make accessible from the other side"`
	}
}

type BonvRequestOptions struct {
	VpcA BonvRequestVpcOptions
	VpcB BonvRequestVpcOptions
}

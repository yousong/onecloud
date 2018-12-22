package utils

import (
	"fmt"

	"yunion.io/x/onecloud/pkg/bonv/cloud/aliyun"
	"yunion.io/x/onecloud/pkg/bonv/cloud/qcloud"
	"yunion.io/x/onecloud/pkg/bonv/cloud/types"
)

func NewClient(provider, account, secret string) (types.Client, error) {
	var client types.Client
	var err error
	switch provider {
	case types.PROVIDER_ALIYUN:
		client, err = aliyun.NewClient(account, secret)
		if err != nil {
			return nil, fmt.Errorf("failed creating aliyun client: %s", err)
		}
	case types.PROVIDER_TENCENTCLOUD:
		client, err = qcloud.NewClient(account, secret)
		if err != nil {
			return nil, fmt.Errorf("failed creating tencentcloud client: %s", err)
		}
	default:
		return nil, fmt.Errorf("account provider not supported yet: %q", provider)
	}
	return client, nil
}

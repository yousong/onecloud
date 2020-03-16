package proxy

import (
	"context"
	"net/http"
	"net/url"

	"golang.org/x/net/http/httpproxy"

	"yunion.io/x/jsonutils"

	proxyapi "yunion.io/x/onecloud/pkg/apis/cloudcommon/proxy"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/mcclient"
)

type SProxySettingManager struct {
	db.SStandaloneResourceBaseManager
}

type SProxySetting struct {
	db.SStandaloneResourceBase

	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string
}

func (man *SProxySettingManager) ValidateCreateData(ctx context.Context, userCred mcclient.TokenCredential, ownerId mcclient.IIdentityProvider, query jsonutils.JSONObject, data proxyapi.ProxySettingCreateInput) (proxyapi.ProxySettingCreateInput, error) {
	return data, nil
}

func (ps *SProxySetting) ValidateUpdateData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data proxyapi.ProxySettingUpdateInput) (proxyapi.ProxySettingUpdateInput, error) {
	return data, nil
}

func (ps *SProxySetting) HttpTransportProxyFunc() proxyapi.HttpTransportProxyFunc {
	cfg := &httpproxy.Config{
		HTTPProxy:  ps.HTTPProxy,
		HTTPSProxy: ps.HTTPSProxy,
		NoProxy:    ps.NoProxy,
	}
	return func(req *http.Request) (*url.URL, error) {
		return cfg.ProxyFunc()(req.URL)
	}
}

package proxy

import (
	"net/http"
	"net/url"

	"yunion.io/x/onecloud/pkg/apis"
)

type HttpTransportProxyFunc func(*http.Request) (*url.URL, error)

type ProxySettingCreateInput struct {
	apis.VirtualResourceCreateInput

	HttpProxy  string
	HttpsProxy string
	NoProxy    string
}

type ProxySettingUpdateInput ProxySettingCreateInput

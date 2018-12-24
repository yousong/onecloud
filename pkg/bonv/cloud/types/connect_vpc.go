package types

type ConnectVpcRequest struct {
	Vpc *Vpc

	AcceptingClient Client
	AcceptingVpc    *Vpc
}

type ConnectVpcResponse struct {
	InitiatingRegionId          string
	InitiatingRouterInterfaceId string

	AcceptingRegionId          string
	AcceptingRouterInterfaceId string
}

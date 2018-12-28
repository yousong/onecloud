package modules

type BonvRequestManager struct {
	ResourceManager
}

var (
	BonvRequests BonvRequestManager
)

func init() {
	BonvRequests = BonvRequestManager{
		NewComputeManager(
			"bonv_request",
			"bonv_requests",
			[]string{
				"id",
			},
			[]string{},
		),
	}
	registerCompute(&BonvRequests)
}

package shell

import (
	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/bonv/shell/modules"
	"yunion.io/x/onecloud/pkg/bonv/shell/options"
)

func init() {
	R(&options.BonvRequestOptions{}, "bonv-request-create", "Create bonv request", func(s *mcclient.ClientSession, opts *options.BonvRequestOptions) error {
		params := jsonutils.Marshal(opts)
		obj, err := modules.BonvRequests.Create(s, params)
		if err != nil {
			return err
		}
		printObject(obj)
		return nil
	})
}

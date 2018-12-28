package shell

import (
	"yunion.io/x/onecloud/pkg/bonv/client/options"
	"yunion.io/x/onecloud/pkg/mcclient"
	mcopts "yunion.io/x/onecloud/pkg/mcclient/options"
)

func init() {
	R(&SubParserArgs{
		Command:     "vpc-import",
		Description: "Import vpc from cloud",
		Opts:        &options.VpcImportOptions{},
		Callback: func(s *mcclient.ClientSession, opts *options.VpcImportOptions) error {
			params, err := mcopts.StructToParams(opts)
			if err != nil {
				return err
			}
			params = params //
			return nil
		},
	})
}

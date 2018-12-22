package service

import (
	"net"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"yunion.io/x/onecloud/pkg/appsrv"
	"yunion.io/x/onecloud/pkg/appsrv/dispatcher"
	"yunion.io/x/onecloud/pkg/bonv/models"
	"yunion.io/x/onecloud/pkg/cloudcommon"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
)

type Options struct {
	Address string `default:"0.0.0.0"`
	Port    int

	WorkerCount int

	cloudcommon.DBOptions
}

func (opts *Options) ListenAddress() string {
	return net.JoinHostPort(opts.Address, strconv.Itoa(opts.Port))
}

func StartService() {
	opts := &Options{
		Address:     "0.0.0.0",
		Port:        8090,
		WorkerCount: 4,

		DBOptions: cloudcommon.DBOptions{
			SqlConnection: "mysql+pymysql://bonv:kQnCdKE49cM=@10.168.222.136:3306/bonv?charset=utf8",
			AutoSyncTable: true,
		},
	}

	appName := "bonv"
	addr := opts.ListenAddress()
	app := appsrv.NewApplication(appName, opts.WorkerCount)
	app.AddHandler("POST", "/cloudconnectrequest", models.HandleNewCloudConnectRequest)
	for _, man := range []db.IModelManager{
		models.CloudConnectManager,
		models.CloudAccountManager,
		models.VpcManager,
	} {
		db.RegisterModelManager(man)
		handler := db.NewModelHandler(man)
		dispatcher.AddModelDispatcher("", app, handler)
	}

	{
		cloudcommon.InitDB(&opts.DBOptions)
		if !db.CheckSync(opts.AutoSyncTable) {
			return
		}
		defer cloudcommon.CloseDB()
	}

	app.ListenAndServe(addr)
}

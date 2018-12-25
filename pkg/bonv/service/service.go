package service

import (
	"net"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"yunion.io/x/log"

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

	cloudcommon.CommonOptions
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

		CommonOptions: cloudcommon.CommonOptions{
			Region:        "Yunion",
			AuthURL:       "http://10.168.222.136:35357/v3",
			AdminUser:     "regionadmin",
			AdminPassword: "GkZOr6poQgcmMszU",
			AdminProject:  "system",
		},
		DBOptions: cloudcommon.DBOptions{
			SqlConnection: "mysql+pymysql://bonv:kQnCdKE49cM=@10.168.222.136:3306/bonv?charset=utf8",
			AutoSyncTable: true,
		},
	}
	commonOpts := &opts.CommonOptions
	dbOpts := &opts.DBOptions

	cloudcommon.InitAuth(commonOpts, func() {
		log.Infof("auth completed")
	})

	appName := "bonv"
	dbAccess := true
	app := appsrv.NewApplication(appName, opts.WorkerCount, dbAccess)
	app.AddHandler("POST", "/cloudconnectrequest", models.HandleNewCloudConnectRequest)
	for _, man := range []db.IModelManager{
		models.CloudConnectManager,
		models.CloudAccountManager,
		models.VpcManager,
		models.VpcConnectManager,
	} {
		db.RegisterModelManager(man)
		handler := db.NewModelHandler(man)
		dispatcher.AddModelDispatcher("", app, handler)
	}

	{
		cloudcommon.InitDB(dbOpts)
		if !db.CheckSync(opts.AutoSyncTable) {
			return
		}
		defer cloudcommon.CloseDB()
	}

	addr := opts.ListenAddress()
	app.ListenAndServe(addr)
}

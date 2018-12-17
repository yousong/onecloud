创建CrossCloudVpcConnection的流程

 1. 选择manager0, 选择cloud0vpc1id, 选择cloud0cidrs, cloud0vpc0id可由平台决定
 2. 同样操作，针对manager1, cloud1vpc1id, cloud1cidrs, cloud1vpc0id
 3. 校验
 	- get or create vpc info
 	- cidrs两个集合不能重叠
 4. 创建CloudPeerConnection
 	- 是否存在，是否正常
 	- cloud0vpc0id, cloud0vpc1id
 	- cloud1vpc0id, cloud1vpc1id
 6. cloud0单臂
     4. 创建CloudPeerConnection
        - 是否存在，是否正常
        - cloud0vpc0id, cloud0vpc1id
     5. cloud0中创建cloudserver
        - cloud0vpc0subnetid中创建cloud0server0
        - cloud0vpc1subnetid中创建cloud0server1
     6. 部署cloud0server0, cloud1server1
     6. /32路由表
        - cloud0vpc0路由表，添加/32条目
        - cloud0vpc1路由表，添加/32条目
        - cloud0vpc1路由表，添加cloud1cidrs条目，指向peer connection
 7. cloud1单臂同


TODO

 option debug, propagate through to remote rpc

提交参数

  POST /crosscloudvpcconnectionrequests

	{
		"request": {
			version: "1",
			clouds: [
				{
					"manager": {
						"provider": "aliyun",
						"access_id": "xxx",
					},
					"vpc": {
						"id": "vpcid0",
						"vswitch_id": "vpcid0vswitchid0",
						"cidrs": [
							"10.60.0.0/16",
						],
					},
				},
				{
					"manager": {
						"provider": "qcloud",
						"access_id": "xxx",
					},
					"vpc": {
						"id": "vpcid1",
						"vswitch_id": "vpcid1vswitchid0",
						"cidrs": [
							"10.62.0.0/16",
						],
					},
				},
			],
		},
	}

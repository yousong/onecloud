// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"net/http"

	"yunion.io/x/onecloud/pkg/multicloud/huawei/client/auth"
	"yunion.io/x/onecloud/pkg/multicloud/huawei/client/auth/credentials"
	"yunion.io/x/onecloud/pkg/multicloud/huawei/client/modules"
)

type Client struct {
	signer    auth.Signer
	regionId  string
	domainId  string
	projectId string

	debug bool

	// 标记初始化状态
	init bool

	Balances           *modules.SBalanceManager
	Bandwidths         *modules.SBandwidthManager
	Disks              *modules.SDiskManager
	Domains            *modules.SDomainManager
	Eips               *modules.SEipManager
	Elasticcache       *modules.SElasticcacheManager
	DcsAvailableZone   *modules.SDcsAvailableZoneManager
	Flavors            *modules.SFlavorManager
	Images             *modules.SImageManager
	OpenStackImages    *modules.SImageManager
	Interface          *modules.SInterfaceManager
	Jobs               *modules.SJobManager
	Keypairs           *modules.SKeypairManager
	Elb                *modules.SLoadbalancerManager
	ElbBackend         *modules.SElbBackendManager
	ElbBackendGroup    *modules.SElbBackendGroupManager
	ElbListeners       *modules.SElbListenersManager
	ElbCertificates    *modules.SElbCertificatesManager
	ElbHealthCheck     *modules.SElbHealthCheckManager
	ElbL7policies      *modules.SElbL7policiesManager
	ElbPolicies        *modules.SElbPoliciesManager
	ElbWhitelist       *modules.SElbWhitelistManager
	Orders             *modules.SOrderManager
	Port               *modules.SPortManager
	Projects           *modules.SProjectManager
	Regions            *modules.SRegionManager
	SecurityGroupRules *modules.SSecgroupRuleManager
	SecurityGroups     *modules.SSecurityGroupManager
	NovaSecurityGroups *modules.SSecurityGroupManager
	Servers            *modules.SServerManager
	ServersV2          *modules.SServerManager
	NovaServers        *modules.SServerManager
	Snapshots          *modules.SSnapshotManager
	OsSnapshots        *modules.SSnapshotManager
	Subnets            *modules.SSubnetManager
	Users              *modules.SUserManager
	Vpcs               *modules.SVpcManager
	Zones              *modules.SZoneManager
	VpcRoutes          *modules.SVpcRouteManager
	SNatRules          *modules.SNatSRuleManager
	DNatRules          *modules.SNatDRuleManager
	NatGateways        *modules.SNatGatewayManager
	DBInstance         *modules.SDBInstanceManager
	DBInstanceBackup   *modules.SDBInstanceBackupManager
	DBInstanceFlavor   *modules.SDBInstanceFlavorManager
	DBInstanceJob      *modules.SDBInstanceJobManager
	Traces             *modules.STraceManager
	CloudEye           *modules.SCloudEyeManager
}

func (self *Client) SetHttpClient(httpClient *http.Client) {
	self.Servers.SetHttpClient(httpClient)
	self.ServersV2.SetHttpClient(httpClient)
	self.NovaServers.SetHttpClient(httpClient)
	self.Snapshots.SetHttpClient(httpClient)
	self.OsSnapshots.SetHttpClient(httpClient)
	self.Images.SetHttpClient(httpClient)
	self.OpenStackImages.SetHttpClient(httpClient)
	self.Projects.SetHttpClient(httpClient)
	self.Regions.SetHttpClient(httpClient)
	self.Zones.SetHttpClient(httpClient)
	self.Vpcs.SetHttpClient(httpClient)
	self.Eips.SetHttpClient(httpClient)
	self.Elasticcache.SetHttpClient(httpClient)
	self.DcsAvailableZone.SetHttpClient(httpClient)
	self.Disks.SetHttpClient(httpClient)
	self.Domains.SetHttpClient(httpClient)
	self.Keypairs.SetHttpClient(httpClient)
	self.Elb.SetHttpClient(httpClient)
	self.ElbBackend.SetHttpClient(httpClient)
	self.ElbBackendGroup.SetHttpClient(httpClient)
	self.ElbListeners.SetHttpClient(httpClient)
	self.ElbCertificates.SetHttpClient(httpClient)
	self.ElbHealthCheck.SetHttpClient(httpClient)
	self.ElbL7policies.SetHttpClient(httpClient)
	self.ElbPolicies.SetHttpClient(httpClient)
	self.ElbWhitelist.SetHttpClient(httpClient)
	self.Orders.SetHttpClient(httpClient)
	self.SecurityGroupRules.SetHttpClient(httpClient)
	self.SecurityGroups.SetHttpClient(httpClient)
	self.NovaSecurityGroups.SetHttpClient(httpClient)
	self.Subnets.SetHttpClient(httpClient)
	self.Users.SetHttpClient(httpClient)
	self.Interface.SetHttpClient(httpClient)
	self.Jobs.SetHttpClient(httpClient)
	self.Balances.SetHttpClient(httpClient)
	self.Bandwidths.SetHttpClient(httpClient)
	self.Port.SetHttpClient(httpClient)
	self.Flavors.SetHttpClient(httpClient)
	self.VpcRoutes.SetHttpClient(httpClient)
	self.SNatRules.SetHttpClient(httpClient)
	self.DNatRules.SetHttpClient(httpClient)
	self.NatGateways.SetHttpClient(httpClient)
	self.DBInstance.SetHttpClient(httpClient)
	self.DBInstanceBackup.SetHttpClient(httpClient)
	self.DBInstanceFlavor.SetHttpClient(httpClient)
	self.DBInstanceJob.SetHttpClient(httpClient)
	self.Traces.SetHttpClient(httpClient)
	self.CloudEye.SetHttpClient(httpClient)
}

func (self *Client) InitWithOptions(regionId, projectId string, credential auth.Credential) error {
	// 从signer中初始化
	signer, err := auth.NewSignerWithCredential(credential)
	if err != nil {
		return err
	}
	self.signer = signer
	self.regionId = regionId
	self.projectId = projectId
	// 暂时还未用到domainId
	self.domainId = ""
	// 初始化 resource manager
	self.initManagers()
	return err
}

func (self *Client) InitWithAccessKey(regionId, projectId, accessKey, secretKey string) error {
	// accessKey signer
	credential := &credentials.AccessKeyCredential{
		AccessKeyId:     accessKey,
		AccessKeySecret: secretKey,
	}

	return self.InitWithOptions(regionId, projectId, credential)
}

func (self *Client) initManagers() {
	if !self.init {
		self.Servers = modules.NewServerManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ServersV2 = modules.NewServerV2Manager(self.regionId, self.projectId, self.signer, self.debug)
		self.NovaServers = modules.NewNovaServerManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Snapshots = modules.NewSnapshotManager(self.regionId, self.projectId, self.signer, self.debug)
		self.OsSnapshots = modules.NewOsSnapshotManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Images = modules.NewImageManager(self.regionId, self.projectId, self.signer, self.debug)
		self.OpenStackImages = modules.NewOpenstackImageManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Projects = modules.NewProjectManager(self.signer, self.debug)
		self.Regions = modules.NewRegionManager(self.signer, self.debug)
		self.Zones = modules.NewZoneManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Vpcs = modules.NewVpcManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Eips = modules.NewEipManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Elasticcache = modules.NewElasticcacheManager(self.regionId, self.projectId, self.signer, self.debug)
		self.DcsAvailableZone = modules.NewDcsAvailableZoneManager(self.regionId, self.signer, self.debug)
		self.Disks = modules.NewDiskManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Domains = modules.NewDomainManager(self.signer, self.debug)
		self.Keypairs = modules.NewKeypairManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Elb = modules.NewLoadbalancerManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ElbBackend = modules.NewElbBackendManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ElbBackendGroup = modules.NewElbBackendGroupManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ElbListeners = modules.NewElbListenersManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ElbCertificates = modules.NewElbCertificatesManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ElbHealthCheck = modules.NewElbHealthCheckManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ElbL7policies = modules.NewElbL7policiesManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ElbPolicies = modules.NewElbPoliciesManager(self.regionId, self.projectId, self.signer, self.debug)
		self.ElbWhitelist = modules.NewElbWhitelistManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Orders = modules.NewOrderManager(self.signer, self.debug)
		self.SecurityGroupRules = modules.NewSecgroupRuleManager(self.regionId, self.projectId, self.signer, self.debug)
		self.SecurityGroups = modules.NewSecurityGroupManager(self.regionId, self.projectId, self.signer, self.debug)
		self.NovaSecurityGroups = modules.NewNovaSecurityGroupManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Subnets = modules.NewSubnetManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Users = modules.NewUserManager(self.signer, self.debug)
		self.Interface = modules.NewInterfaceManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Jobs = modules.NewJobManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Balances = modules.NewBalanceManager(self.signer, self.debug)
		self.Bandwidths = modules.NewBandwidthManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Port = modules.NewPortManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Flavors = modules.NewFlavorManager(self.regionId, self.projectId, self.signer, self.debug)
		self.VpcRoutes = modules.NewVpcRouteManager(self.regionId, self.projectId, self.signer, self.debug)
		self.SNatRules = modules.NewNatSManager(self.regionId, self.projectId, self.signer, self.debug)
		self.DNatRules = modules.NewNatDManager(self.regionId, self.projectId, self.signer, self.debug)
		self.NatGateways = modules.NewNatGatewayManager(self.regionId, self.projectId, self.signer, self.debug)
		self.DBInstance = modules.NewDBInstanceManager(self.regionId, self.projectId, self.signer, self.debug)
		self.DBInstanceBackup = modules.NewDBInstanceBackupManager(self.regionId, self.projectId, self.signer, self.debug)
		self.DBInstanceFlavor = modules.NewDBInstanceFlavorManager(self.regionId, self.projectId, self.signer, self.debug)
		self.DBInstanceJob = modules.NewDBInstanceJobManager(self.regionId, self.projectId, self.signer, self.debug)
		self.Traces = modules.NewTraceManager(self.regionId, self.projectId, self.signer, self.debug)
		self.CloudEye = modules.NewCloudEyeManager(self.regionId, self.projectId, self.signer, self.debug)
	}

	self.init = true
}

func NewClientWithAccessKey(regionId, projectId, accessKey, secretKey string, debug bool) (*Client, error) {
	c := &Client{debug: debug}
	err := c.InitWithAccessKey(regionId, projectId, accessKey, secretKey)
	return c, err
}

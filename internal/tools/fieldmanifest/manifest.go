// Copyright 2022-2025 The sacloud/iaas-api-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fieldmanifest holds an allowlist of fields that should be emitted
// into the generated TypeSpec (and thus the v2 OpenAPI / v2 Go client).
//
// downstream (usacloud / terraform-provider-sakuracloud / terraform-provider-sakura)
// は iaas-api-go の一部のフィールドしか参照しない。v2 面積を最小化するため、
// 実際に使われているフィールドだけを TypeSpec に emit し、それ以外は生成時に落とす。
//
// 登録されていないモデルはデフォルトで全フィールドを emit する (pass-through)。
// 段階導入のため、未整理のリソースは従来どおりの挙動を維持する。
//
// キーは TypeSpec モデル名 (合成サブモデル名を含む)。値は TypeSpec フィールド名の
// セット。mapconv でリネームされる場合は TypeSpec 側の名前 (例: v1 `IconID` →
// v2 `Icon`) を使う。
//
// 関連: internal/tools/findmanifest/manifest.go (q クエリパラメータの allowlist)。
// 責務が直交するので意図的に別パッケージに分けている。
package fieldmanifest

// Manifest[modelName][fieldName] = true のとき、そのフィールドは emit される。
// modelName 自体が Manifest に存在しないときは従来通り全フィールドが emit される。
var Manifest = map[string]map[string]bool{
	// ----- Archive -----
	// v2 terraform-provider-sakuracloud: resource_sakuracloud_archive.go,
	//   data_source_sakuracloud_archive.go, resource_sakuracloud_archive_share.go
	// v3 terraform-provider-sakura: internal/service/archive/resource.go, data_source.go
	// iaas-service-go: archive/builder/*, archive/{read,find,update}_service.go
	"Archive": {
		"ID":           true, // downstream Read / SetId
		"Name":         true, // downstream Read / Update
		"Description":  true, // downstream Read / Update
		"Tags":         true, // downstream Read / Update
		"Icon":         true, // downstream Read (IconID via mapconv)
		"SizeMB":       true, // downstream Read (GetSizeGB)
		"Availability": true, // v2 archive_share.go: Availability.IsUploading()
	},
	// iaas-service-go/archive/builder/{standard,blank,transfer,from_shared}_archive_builder.go
	// で設定されるフィールドを union したもの。v2 expandArchiveBuilder / v3 同等経路。
	"ArchiveCreateRequest": {
		"Name":          true,
		"Description":   true,
		"Tags":          true,
		"Icon":          true, // IconID
		"SizeMB":        true, // blank builder (SizeGB * GiB)
		"SourceDisk":    true, // SourceDiskID
		"SourceArchive": true, // SourceArchiveID
	},
	// v2 expandArchiveUpdateRequest / v3 同等経路。
	"ArchiveUpdateRequest": {
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
	},

	// ----- Switch -----
	// v2 terraform-provider-sakuracloud: resource_sakuracloud_switch.go,
	//   data_source_sakuracloud_switch.go, resource_sakuracloud_internet.go (Subnets),
	//   data_source_sakuracloud_subnet.go (Subnets)
	// v3 terraform-provider-sakura: internal/service/switch/*
	// iaas-service-go: swytch/*
	"Switch": {
		"ID":          true,
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
		"ServerCount": true, // v2 resource_switch.go / resource_internet.go
		"Bridge":      true, // v2 resource_switch.go: BridgeID 参照
		"Subnets":     true, // v2 data_source_subnet / resource_internet
	},
	"SwitchCreateRequest": {
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
	},
	"SwitchUpdateRequest": {
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
	},

	// ----- Bridge -----
	// v2 terraform-provider-sakuracloud: resource_sakuracloud_bridge.go,
	//   data_source_sakuracloud_bridge.go
	// v3 terraform-provider-sakura: internal/service/bridge/*
	"Bridge": {
		"ID":          true,
		"Name":        true,
		"Description": true,
	},
	"BridgeCreateRequest": {
		"Name":        true,
		"Description": true,
	},
	"BridgeUpdateRequest": {
		"Name":        true,
		"Description": true,
	},

	// ----- CDROM -----
	// v2: resource_sakuracloud_cdrom.go, data_source_sakuracloud_cdrom.go, structure_cdrom.go
	// iaas-service-go: cdrom/*
	"CDROM": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true, // IconID
		"Availability": true, // FTP アップロード中かどうかの判定 (iaas-service-go 同等の用途)
	},
	"CDROMCreateRequest": {
		"SizeMB":      true,
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
	},
	"CDROMUpdateRequest": {
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
	},

	// ----- Icon -----
	// v2: resource_sakuracloud_icon.go, data_source_sakuracloud_icon.go, structure_icon.go
	// v3: internal/service/icon/*
	"Icon": {
		"ID":   true,
		"Name": true,
		"Tags": true,
		"URL":  true,
	},
	"IconCreateRequest": {
		"Name":  true,
		"Tags":  true,
		"Image": true,
	},
	"IconUpdateRequest": {
		"Name": true,
		"Tags": true,
	},

	// ----- Note -----
	// v2: resource_sakuracloud_note.go, data_source_sakuracloud_note.go, structure_note.go
	// v3: internal/service/script/* (iaas-api-go Note は sakura プロバイダでは "script" と呼ばれる)
	"Note": {
		"ID":          true,
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Class":       true,
		"Content":     true,
		"Icon":        true,
	},
	"NoteCreateRequest": {
		"Name":    true,
		"Tags":    true,
		"Icon":    true,
		"Class":   true,
		"Content": true,
	},
	"NoteUpdateRequest": {
		"Name":    true,
		"Tags":    true,
		"Icon":    true,
		"Class":   true,
		"Content": true,
	},

	// ----- SSHKey -----
	// v2: resource_sakuracloud_ssh_key.go, structure_ssh_key.go
	// v3: internal/service/ssh_key/*
	"SSHKey": {
		"ID":          true,
		"Name":        true,
		"Description": true,
		"PublicKey":   true,
		"Fingerprint": true,
	},
	"SSHKeyCreateRequest": {
		"Name":        true,
		"Description": true,
		"PublicKey":   true,
	},
	"SSHKeyUpdateRequest": {
		"Name":        true,
		"Description": true,
	},

	// ----- Disk -----
	// v2: resource_sakuracloud_disk.go, data_source_sakuracloud_disk.go, structure_disk.go
	// v3: internal/service/disk/*
	// iaas-service-go: disk/* (builder 経由で terraform v2/v3 が使用)
	"Disk": {
		"ID":                  true,
		"Name":                true,
		"Description":         true,
		"Tags":                true,
		"Connection":          true,
		"SizeMB":              true,
		"Plan":                true, // DiskPlanID
		"Server":              true, // ServerID
		"SourceArchive":       true, // SourceArchiveID
		"SourceDisk":          true, // SourceDiskID
		"Icon":                true,
		"EncryptionAlgorithm": true,
		"EncryptionKey":       true, // KMSKeyID
		"Storage":             true, // v3: DedicatedStorageContractID
		"Availability":        true, // iaas-service-go/disk/update_request.go が Read の "Availabilities.Available" 前提チェックに使用
	},
	"DiskCreateRequest": {
		"Plan":                true,
		"Connection":          true,
		"SourceDisk":          true,
		"SourceArchive":       true,
		"SizeMB":              true,
		"Name":                true,
		"Description":         true,
		"Tags":                true,
		"Icon":                true,
		"EncryptionAlgorithm": true,
	},
	"DiskUpdateRequest": {
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
		"Connection":  true,
	},

	// ----- Server -----
	// v2: resource_sakuracloud_server.go, structure_server.go
	// v3: internal/service/server/*
	// iaas-service-go: server/* (builder, disk 接続など)
	"Server": {
		"ID":              true,
		"Name":            true,
		"Description":     true,
		"Tags":            true,
		"HostName":        true,
		"InterfaceDriver": true,
		"ServerPlan":      true, // CPU/MemoryMB/GPU/GPUModel/CPUModel/Commitment
		"Zone":            true, // Region.NameServers 経由
		"Instance":        true, // CDROM.ID (CDROMID) 参照
		"Disks":           true, // Connected disks
		"Interfaces":      true, // NIC 情報
		"PrivateHost":     true, // ID/Name
		"Icon":            true,
		"Availability":    true, // iaas-service-go/server/update_request.go が Update 前チェックに使用
	},
	"ServerCreateRequest": {
		"ServerPlan":        true,
		"ConnectedSwitches": true,
		"InterfaceDriver":   true,
		"Name":              true,
		"Description":       true,
		"Tags":              true,
		"Icon":              true,
		"PrivateHost":       true,
	},
	"ServerUpdateRequest": {
		"Name":            true,
		"Description":     true,
		"Tags":            true,
		"Icon":            true,
		"PrivateHost":     true,
		"InterfaceDriver": true,
	},

	// ----- Internet -----
	// v2: resource_sakuracloud_internet.go
	// v3: internal/service/internet/*
	"Internet": {
		"ID":             true,
		"Name":           true,
		"Description":    true,
		"Tags":           true,
		"Icon":           true,
		"BandWidthMbps":  true,
		"NetworkMaskLen": true,
		"Switch":         true, // SwitchInfo (ID, Subnets[0], IPv6Nets[0], ServerCount)
	},
	"InternetCreateRequest": {
		"Name":           true,
		"Description":    true,
		"Tags":           true,
		"Icon":           true,
		"NetworkMaskLen": true,
		"BandWidthMbps":  true,
	},
	"InternetUpdateRequest": {
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
	},

	// ----- PacketFilter -----
	// v2: resource_sakuracloud_packet_filter.go, structure_packet_filter.go
	// v3: internal/service/packet_filter/*
	"PacketFilter": {
		"ID":          true,
		"Name":        true,
		"Description": true,
		"Expression":  true,
	},
	"PacketFilterCreateRequest": {
		"Name":        true,
		"Description": true,
		"Expression":  true,
	},
	"PacketFilterUpdateRequest": {
		"Name":        true,
		"Description": true,
		"Expression":  true,
	},

	// ----- PrivateHost -----
	// v2: resource_sakuracloud_private_host.go, data_source_sakuracloud_private_host.go
	// v3: internal/service/private_host/*
	"PrivateHost": {
		"ID":               true,
		"Name":             true,
		"Description":      true,
		"Tags":             true,
		"Icon":             true,
		"Plan":             true, // PlanClass / PlanID
		"AssignedCPU":      true,
		"AssignedMemoryMB": true,
		"Host":             true, // HostName 経由
	},
	"PrivateHostCreateRequest": {
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
		"Plan":        true,
	},
	"PrivateHostUpdateRequest": {
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
	},

	// ----- Subnet -----
	// v2: data_source_sakuracloud_subnet.go
	// v3: internal/service/subnet/*
	// Subnet は Internet から AddSubnet / UpdateSubnet 経由で作成する (独立 Create API なし)。
	// response のみ allowlist 対象。Create/Update は subnet 用の独自 API ではなく internet 側にある。
	"Subnet": {
		"ID":             true,
		"Switch":         true, // SubnetSwitch (ID, Internet)
		"NextHop":        true,
		"NetworkAddress": true,
		"NetworkMaskLen": true,
		"IPAddresses":    true, // Min/Max 抽出用
	},

	// ========== Appliance 系 (fat model) ==========
	// Appliance は /api/cloud/1.1/appliance 共有エンドポイントで Database が代表レスポンス型。
	// ただし v1 Go 型は variant ごとに独立 (iaas.Database / iaas.LoadBalancer / iaas.NFS 等)。
	// 各 variant の top-level response フィールドはほぼ全て downstream で使われるため、
	// 主な prune 対象は CreatedAt / ModifiedAt 程度。サブモデル (Settings, Remark, Instance etc.)
	// は現時点では pass-through (未登録 = 全通過) にしておく。
	// 深いネストの絞込みは将来タスク。

	// ----- Database -----
	// v2 resource_sakuracloud_database*.go / v3 internal/service/database/*
	// iaas-service-go/database/builder/*
	"Database": {
		"ID":                true,
		"Class":             true,
		"Name":              true,
		"Description":       true,
		"Tags":              true,
		"Availability":      true, // iaas-service-go/database/update_request.go
		"Icon":              true,
		"Settings":          true,
		"InterfaceSettings": true,
		"SettingsHash":      true,
		"Instance":          true,
		"Remark":            true,
		"IPAddresses":       true,
		"Disk":              true,
	},

	// ----- LoadBalancer -----
	// v2 resource_sakuracloud_load_balancer.go / v3 enhanced_lb は別系統 (ProxyLB 系)
	// iaas-service-go/loadbalancer/builder/*
	"LoadBalancer": {
		"ID":                 true,
		"Name":               true,
		"Description":        true,
		"Tags":               true,
		"Availability":       true,
		"Class":              true,
		"Icon":               true,
		"Instance":           true,
		"Remark":             true,
		"IPAddresses":        true,
		"VirtualIPAddresses": true,
		"SettingsHash":       true,
		"Interfaces":         true,
	},

	// ----- NFS -----
	// v2 resource_sakuracloud_nfs.go / v3 internal/service/nfs/*
	// iaas-service-go/nfs/builder/*
	"NFS": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Availability": true,
		"Class":        true,
		"Instance":     true,
		"Interfaces":   true,
		"Remark":       true,
		"IPAddresses":  true,
		"Icon":         true,
		"Switch":       true,
	},

	// ----- VPCRouter -----
	// v2 resource_sakuracloud_vpc_router.go / v3 internal/service/vpn_router/*
	// iaas-service-go/vpcrouter/builder/*
	"VPCRouter": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Availability": true,
		"Class":        true,
		"Icon":         true,
		"Instance":     true,
		"Remark":       true,
		"Settings":     true,
		"SettingsHash": true,
		"Interfaces":   true,
	},

	// ----- MobileGateway -----
	// v2 resource_sakuracloud_mobile_gateway.go / v3 は未対応
	// iaas-service-go/mobilegateway/builder/*
	"MobileGateway": {
		"ID":                true,
		"Name":              true,
		"Description":       true,
		"Tags":              true,
		"Availability":      true,
		"Class":             true,
		"Icon":              true,
		"Instance":          true,
		"Remark":            true,
		"Settings":          true,
		"SettingsHash":      true,
		"Interfaces":        true,
		"InterfaceSettings": true,
		"StaticRoutes":      true,
		"SIMRoutes":         true,
	},

	// ========== CommonServiceItem 系 ==========
	// 共有エンドポイント /api/cloud/1.1/commonserviceitem。variant ごとに独立した Go 型。

	// ----- DNS -----
	// v2 resource_sakuracloud_dns.go
	"DNS": {
		"ID":           true,
		"Name":         true, // DNS の ZoneName (= Name) として使用
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"Records":      true,
		"Settings":     true,
		"SettingsHash": true,
		"Status":       true, // Zone / NS[]
	},

	// ----- GSLB -----
	// v2 resource_sakuracloud_gslb.go
	"GSLB": {
		"ID":                 true,
		"Name":               true,
		"Description":        true,
		"Tags":               true,
		"Icon":               true,
		"SettingsHash":       true,
		"Status":             true, // FQDN
		"Settings":           true,
		"DestinationServers": true,
	},

	// ----- ProxyLB -----
	// v2 resource_sakuracloud_proxy_lb.go / v3 enhanced_lb (ProxyLB と同系統)
	"ProxyLB": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"Plan":         true,
		"Settings":     true,
		"BindPorts":    true,
		"Servers":      true,
		"Rules":        true,
		"SettingsHash": true,
		"Status":       true, // UseVIPFailover / Region / FQDN / VirtualIPAddress
	},

	// ----- AutoBackup -----
	// v2 resource_sakuracloud_auto_backup.go
	"AutoBackup": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"Settings":     true,
		"SettingsHash": true,
		"Status":       true, // DiskID
	},

	// ----- AutoScale -----
	// v2 resource_sakuracloud_auto_scale.go
	"AutoScale": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"Settings":     true,
		"SettingsHash": true,
		"Status":       true,
	},

	// ----- CertificateAuthority -----
	// v2 resource_sakuracloud_certificate_authority.go
	"CertificateAuthority": {
		"ID":          true,
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
		"Status":      true,
	},

	// ----- ContainerRegistry -----
	// v2 resource_sakuracloud_container_registry.go / v3 internal/service/container_registry/*
	"ContainerRegistry": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"Settings":     true,
		"SettingsHash": true,
		"Status":       true, // SubDomainLabel, FQDN など
	},

	// ----- EnhancedDB -----
	// v2 resource_sakuracloud_enhanced_db.go
	"EnhancedDB": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"SettingsHash": true,
		"Status":       true, // DatabaseName, DatabaseType, Region, HostName 等
	},

	// ----- ESME -----
	// v2 resource_sakuracloud_esme.go
	"ESME": {
		"ID":          true,
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Icon":        true,
	},

	// ----- LocalRouter -----
	// v2 resource_sakuracloud_local_router.go
	"LocalRouter": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Availability": true, // v2: data.Availability.IsFailed() チェックあり
		"Icon":         true,
		"Settings":     true,
		"Peers":        true,
		"StaticRoutes": true,
		"SettingsHash": true,
		"Status":       true,
	},

	// ----- SIM -----
	// v2 resource_sakuracloud_sim.go
	"SIM": {
		"ID":          true,
		"Name":        true,
		"Description": true,
		"Tags":        true,
		"Class":       true,
		"Status":      true,
		"Icon":        true,
	},

	// ----- SimpleMonitor -----
	// v2 resource_sakuracloud_simple_monitor.go
	"SimpleMonitor": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"Class":        true,
		"Status":       true,
		"Settings":     true,
		"SettingsHash": true,
	},

	// ----- SimpleNotificationDestination -----
	// v3 internal/service/simple_notification_destination/*
	"SimpleNotificationDestination": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"Settings":     true,
		"SettingsHash": true,
	},

	// ----- SimpleNotificationGroup -----
	// v3 internal/service/simple_notification_group/*
	"SimpleNotificationGroup": {
		"ID":           true,
		"Name":         true,
		"Description":  true,
		"Tags":         true,
		"Icon":         true,
		"Settings":     true,
		"SettingsHash": true,
	},

	// ========== Plan / Facility / Misc ==========

	// ----- DiskPlan -----
	// 全フィールド使用 (usacloud bills / terraform plan lookup)。Size の sub-model は pass-through。
	"DiskPlan": {
		"ID":           true,
		"Name":         true,
		"StorageClass": true,
		"Availability": true,
		"Size":         true,
	},

	// ----- InternetPlan -----
	"InternetPlan": {
		"ID":            true,
		"Name":          true,
		"BandWidthMbps": true,
		"Availability":  true,
	},

	// ----- ServerPlan -----
	// ConfidentialVM は usacloud/terraform 共に未使用。
	"ServerPlan": {
		"ID":           true,
		"Name":         true,
		"CPU":          true,
		"MemoryMB":     true,
		"GPU":          true,
		"GPUModel":     true,
		"CPUModel":     true,
		"Commitment":   true,
		"Generation":   true,
		"Availability": true,
	},

	"PrivateHostPlan": {
		"ID":           true,
		"Name":         true,
		"Class":        true,
		"CPU":          true,
		"MemoryMB":     true,
		"Availability": true,
		"Dedicated":    true, // v2/integration private_host_test.go で参照
	},

	// ----- LicenseInfo -----
	// CreatedAt/ModifiedAt は未使用 (TermsOfUse は usacloud で参照)。
	"LicenseInfo": {
		"ID":         true,
		"Name":       true,
		"TermsOfUse": true,
	},

	// ----- ServiceClass -----
	// usacloud のみで使用 (CLI 出力)。
	"ServiceClass": {
		"ID":               true,
		"ServiceClassName": true,
		"ServiceClassPath": true,
		"DisplayName":      true,
		"IsPublic":         true,
		"Price":            true,
	},

	// ----- License -----
	// terraform v2/v3 は未実装。usacloud のみ使用。
	"License": {
		"ID":          true,
		"Name":        true,
		"LicenseInfo": true,
	},

	// ----- Zone -----
	// DisplayOrder, IsDummy, VNCProxy, FTPServer は未使用。
	// Zone.Region は Server.Zone.Region.NameServers 経由で使用されるため含める。
	"Zone": {
		"ID":          true,
		"Name":        true,
		"Description": true,
		"Region":      true,
	},

	// ----- IPAddress -----
	// Interface, Subnet は未使用 (v2/v3 ipv4_ptr resource は IPAddress/HostName のみ使用)。
	"IPAddress": {
		"IPAddress": true,
		"HostName":  true,
	},

	// ----- IPv6Addr -----
	// v1 usacloud のみ使用。IPv6Net, Interface は未使用。
	"IPv6Addr": {
		"IPv6Addr": true,
		"HostName": true,
	},

	// ----- IPv6Net -----
	// usacloud のみ使用。v2/integration/internet_test.go が ID を EnableIPv6 後に読む。
	"IPv6Net": {
		"ID":            true, // v2/integration internet EnableIPv6 後の参照
		"IPv6Prefix":    true,
		"IPv6PrefixLen": true,
	},

	// ----- Interface -----
	// terraform は Interface CRUD を直接叩かないが、v2/integration/interface_test.go が
	// Create/Read/Update/ConnectToSwitch/DisconnectFromSwitch をテストする。
	// Server.Interfaces[] の表示は別モデル InterfaceView を使う (Interface とは別構造)。
	"Interface": {
		"ID":            true,
		"Server":        true,
		"Switch":        true, // ConnectToSwitch 後の確認
		"UserIPAddress": true, // Update の結果確認
	},

	// ----- Bill -----
	// usacloud only。MemberID, PayLimit, PaymentClassID は未使用。
	"Bill": {
		"ID":     true,
		"Amount": true,
		"Date":   true,
		"Paid":   true,
	},

	// ----- Coupon -----
	// usacloud only。MemberID, ContractID, ServiceClassID は未使用。
	"Coupon": {
		"ID":        true,
		"Discount":  true,
		"AppliedAt": true,
		"UntilAt":   true,
	},

	// ----- AuthStatus -----
	// iaas-service-go が Account ID / Member Code / IsAPIKey を利用する。
	// AuthClass, AuthMethod, ExternalPermission, OperationPenalty, Permission は未使用。
	"AuthStatus": {
		"Account":  true,
		"Member":   true,
		"IsAPIKey": true,
	},

	// ========== Sub-model pruning ==========
	// Appliance 等の Interface 関連サブモデルで、downstream で参照されないフィールドを除外する。

	// ----- InterfaceView -----
	// Server.Interfaces[] / Database.Interfaces[] 等で使われる共有モデル。
	// HostName は v2/v3/iaas-service-go いずれも未参照 (getter のみ zz_models.go に存在)。
	"InterfaceView": {
		"ID":            true,
		"IPAddress":     true,
		"MACAddress":    true, // v2 structure_server.go
		"UserIPAddress": true, // v2/iaas-service-go
		"Switch":        true, // Switch.ID / Switch.Scope (SwitchID/SwitchScope accessor)
		"PacketFilter":  true, // PacketFilter.ID (PacketFilterID accessor)
	},

	// ----- InterfaceViewPacketFilter -----
	// downstream は ID のみ読む。Name は未参照、RequiredHostVersionn は DSL 側の typo (double n)。
	"InterfaceViewPacketFilter": {
		"ID": true,
	},

	// ----- VPCRouterInterface -----
	// v2/v3 structure_vpc_router.go は IPAddress / SwitchID / SubnetNetworkMaskLen 経由で参照。
	// iaas-service-go builder は iface.Index (modelFieldExclusions 済) と iface.SwitchID のみ参照。
	// MACAddress, UserIPAddress, HostName は未参照。
	"VPCRouterInterface": {
		"ID":           true,
		"IPAddress":    true,
		"Switch":       true,
		"PacketFilter": true,
	},

	// ----- VPCRouterInterfacePacketFilter -----
	// 同じく ID のみ。
	"VPCRouterInterfacePacketFilter": {
		"ID": true,
	},

	// ----- DatabaseSettingBackup (legacy backup) -----
	// Backup (v1) は Rotate/Time/DayOfWeek のみ使用。Connect は Backupv2 側のみで指定される
	// (iaas-service-go/database/apply_request.go: Connect は Backupv2 にだけ set される)。
	"DatabaseSettingBackup": {
		"Rotate":    true,
		"Time":      true,
		"DayOfWeek": true,
	},

	// ----- DatabaseSettingBackupv2View (response) -----
	// FirstEnabledAt は zz_models.go の getter/setter にのみ存在、downstream 未参照。
	"DatabaseSettingBackupv2View": {
		"Rotate":    true,
		"Time":      true,
		"DayOfWeek": true,
		"Connect":   true,
	},

	// ----- DatabaseRemarkDBConfCommon -----
	// 同一モデルが response (Remark) と request (CreateRequestRemark.DBConf.Common) の両方で使われる。
	// response 側は DatabaseName/Version/Revision、Create 側は DefaultUser/UserPassword/DatabaseName を利用
	// (v2/integration/database_appliance_test.go)。union を allowlist に持つ。

	// ----- AutoBackupStatus -----
	// downstream は DiskID のみ参照 (v2 structure_auto_backup.go: data.DiskID)。
	// AccountID/ZoneID/ZoneName は参照なし。
	"AutoBackupStatus": {
		"DiskID": true,
	},
}

// IsRegistered は modelName が allowlist を持つかを返す。
// false の場合、呼び出し側は全フィールドを emit する (pass-through)。
func IsRegistered(modelName string) bool {
	_, ok := Manifest[modelName]
	return ok
}

// Allows は (modelName, fieldName) が emit 対象かどうかを返す。
// modelName が未登録の場合は常に true (pass-through)。
func Allows(modelName, fieldName string) bool {
	allow, ok := Manifest[modelName]
	if !ok {
		return true
	}
	return allow[fieldName]
}

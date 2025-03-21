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

package fake

import (
	"context"

	"github.com/sacloud/iaas-api-go"
)

// List is fake implementation
func (o *IPAddressOp) List(ctx context.Context, zone string) (*iaas.IPAddressListResult, error) {
	return &iaas.IPAddressListResult{
		Total: 1,
		Count: 1,
		From:  0,
		IPAddress: []*iaas.IPAddress{
			{
				HostName:  "",
				IPAddress: "192.0.2.1",
			},
		},
	}, nil
}

// Read is fake implementation
func (o *IPAddressOp) Read(ctx context.Context, zone string, ipAddress string) (*iaas.IPAddress, error) {
	return &iaas.IPAddress{
		HostName:  "",
		IPAddress: ipAddress,
	}, nil
}

// UpdateHostName is fake implementation
func (o *IPAddressOp) UpdateHostName(ctx context.Context, zone string, ipAddress string, hostName string) (*iaas.IPAddress, error) {
	return &iaas.IPAddress{
		HostName:  hostName,
		IPAddress: ipAddress,
	}, nil
}

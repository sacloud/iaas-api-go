// Copyright 2022 The sacloud/iaas-api-go Authors
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
	"errors"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

// List is fake implementation
func (o *WebAccelOp) List(ctx context.Context) (*iaas.WebAccelListResult, error) {
	results, _ := find(o.key, iaas.APIDefaultZone, nil)
	var values []*iaas.WebAccel
	for _, res := range results {
		dest := &iaas.WebAccel{}
		copySameNameField(res, dest)
		values = append(values, dest)
	}
	return &iaas.WebAccelListResult{
		Total:     len(results),
		Count:     len(results),
		From:      0,
		WebAccels: values,
	}, nil
}

// Read is fake implementation
func (o *WebAccelOp) Read(ctx context.Context, id types.ID) (*iaas.WebAccel, error) {
	value := getWebAccelByID(iaas.APIDefaultZone, id)
	if value == nil {
		return nil, newErrorNotFound(o.key, id)
	}
	dest := &iaas.WebAccel{}
	copySameNameField(value, dest)
	return dest, nil
}

// ReadCertificate is fake implementation
func (o *WebAccelOp) ReadCertificate(ctx context.Context, id types.ID) (*iaas.WebAccelCerts, error) {
	// valid only when running acc test
	err := errors.New("not implements")
	return nil, err
}

// CreateCertificate is fake implementation
func (o *WebAccelOp) CreateCertificate(ctx context.Context, id types.ID, param *iaas.WebAccelCertRequest) (*iaas.WebAccelCerts, error) {
	// valid only when running acc test
	err := errors.New("not implements")
	return nil, err
}

// UpdateCertificate is fake implementation
func (o *WebAccelOp) UpdateCertificate(ctx context.Context, id types.ID, param *iaas.WebAccelCertRequest) (*iaas.WebAccelCerts, error) {
	// valid only when running acc test
	err := errors.New("not implements")
	return nil, err
}

// DeleteCertificate is fake implementation
func (o *WebAccelOp) DeleteCertificate(ctx context.Context, id types.ID) error {
	return errors.New("not implements")
}

// DeleteAllCache is fake implementation
func (o *WebAccelOp) DeleteAllCache(ctx context.Context, param *iaas.WebAccelDeleteAllCacheRequest) error {
	return nil
}

// DeleteCache is fake implementation
func (o *WebAccelOp) DeleteCache(ctx context.Context, param *iaas.WebAccelDeleteCacheRequest) ([]*iaas.WebAccelDeleteCacheResult, error) {
	var result []*iaas.WebAccelDeleteCacheResult
	for _, url := range param.URL {
		result = append(result, &iaas.WebAccelDeleteCacheResult{
			URL:    url,
			Status: 404,
			Result: "Not Found",
		})
	}
	return result, nil
}

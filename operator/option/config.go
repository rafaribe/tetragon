// Copyright 2020 Authors of Cilium
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

package option

import (
	"github.com/spf13/viper"
)

const (
	// SkipCRDCreation specifies whether the CustomResourceDefinition will be
	// disabled for the operator
	SkipCRDCreation = "skip-crd-creation"

	// CMDRef is the path to cmdref output directory
	CMDRef = "cmdref"
)

// OperatorConfig is the configuration used by the operator.
type OperatorConfig struct {
	// SkipCRDCreation disables creation of the CustomResourceDefinition
	// for the operator
	SkipCRDCreation bool
}

// Populate sets all options with the values from viper.
func (c *OperatorConfig) Populate() {
	c.SkipCRDCreation = viper.GetBool(SkipCRDCreation)
}

// Config represents the operator configuration.
var Config = &OperatorConfig{}

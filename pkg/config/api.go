// Copyright 2017 Google Inc.
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

package config

// APIMethod constants are used to refer to the methods handled by api.Handler
type APIMethod int

const (
	// CheckMethod represents Check operation of the mixer api.
	CheckMethod APIMethod = iota
	// ReportMethod represents Report operation of the mixer api.
	ReportMethod
	// QuotaMethod represents Quota operation of the mixer api.
	QuotaMethod
)

// APIMethods is the authoritative place that lists supported api methods.
var APIMethods = []APIMethod{CheckMethod, ReportMethod, QuotaMethod}

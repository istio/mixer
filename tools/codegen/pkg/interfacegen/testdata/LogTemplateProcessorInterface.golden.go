// Copyright 2017 Istio Authors
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

// THIS FILE IS AUTOMATICALLY GENERATED.

package istio_mixer_adapter_log

import (
	"istio.io/mixer/pkg/adapter/config"
)

// Fully qualified name of this template
const TemplateName = "istio.mixer.adapter.log.Log"

// Instance is constructed by Mixer for 'istio.mixer.adapter.log.Log' template.
//
// Defines the format of a single log entry.
// example :
// ...
// ...
type Instance struct {
	Name string

	// Defines the format of a single log entry.
	// example :
	// ...
	// ...
	Dimensions map[string]interface{}
}

// LogHandlerBuilder must be implemented by adapter code if it wants to
// process data associated with the template.
//
// Using this interface, during configuration phase, Mixer
// will call into the adapter to configure it with adapter specific configuration
// as well as all inferred types.
type LogHandlerBuilder interface {
	config.HandlerBuilder
	// ConfigureLogHandler is invoked by Mixer to pass all possible Types for this template to the adapter.
	// Type hold information about the shape of the Instances that will be dispatched to the
	// adapters at request time. Adapter can expect to receive corresponding Instance objects at request time.
	ConfigureLogHandler(map[string]*Type /*Instance name -> Type*/) error
}

// LogHandler must be implemented by adapter code if it wants to
// process data associated with the template.
//
// Using this interface, during request-time, Mixer
// Mixer dispatches the created instances (based on request time attribute and operator-supplied configuration to map
// attributes into template specific instances) to the adapters. Adapters take the incoming instances and do what they
// need to achieve their primary function.
type LogHandler interface {
	config.Handler
	HandleLog([]*Instance) error
}

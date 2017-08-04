// Copyright 2017 Istio Authors.
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

package aspect

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	rpc "github.com/googleapis/googleapis/google/rpc"

	"istio.io/mixer/pkg/adapter"
	config2 "istio.io/mixer/pkg/adapter/config"
	"istio.io/mixer/pkg/attribute"
	"istio.io/mixer/pkg/config"
	"istio.io/mixer/pkg/config/descriptor"
	cpb "istio.io/mixer/pkg/config/proto"
	"istio.io/mixer/pkg/expr"
	"istio.io/mixer/pkg/template"
)

type (
	reportManager struct {
		repo template.Repository
	}

	reportExecutor struct {
		tmplName     string
		procDispatch template.ProcessReportFn
		hndlr        config2.Handler
		ctrs         map[string]proto.Message // constructor name -> constructor params
	}
)

// NewReportManager creates a ReportManager. TODO make this non public once adapterManager starts using this.
// For now, made it public to please the linter for unused fn error.
func NewReportManager(repo template.Repository) ReportManager {
	return &reportManager{repo: repo}
}

func (m *reportManager) NewReportExecutor(c *cpb.Combined, createAspect CreateAspectFunc, env adapter.Env,
	df descriptor.Finder, tmpl string) (ReportExecutor, error) {
	ctrs := make(map[string]proto.Message)
	for _, cstr := range c.Constructors {
		ctrs[cstr.Name] = cstr.Params.(proto.Message)
		if cstr.Template != tmpl {
			return nil, fmt.Errorf("resolved constructor's '%v' template is different than expected template name : %s", cstr, tmpl)
		}
	}

	out, err := createAspect(env, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct report aspect with config '%v': %v", c, err)
	}

	// adapter.Aspect is identical to adapter.config.Handler, this cast has to pass.
	v, _ := out.(config2.Handler)

	ti, _ := m.repo.GetTemplateInfo(tmpl)
	if b := ti.HandlerSupportsTemplate(v); !b {
		return nil, fmt.Errorf("Handler does not implement interface %s. "+
			"Therefore, it cannot support template %v", ti.HndlrName, tmpl)
	}

	return &reportExecutor{tmpl, ti.ProcessReport, v, ctrs}, nil
}

func (*reportManager) DefaultConfig() config.AspectParams { return nil }
func (*reportManager) ValidateConfig(c config.AspectParams, tc expr.TypeChecker, df descriptor.Finder) (ce *adapter.ConfigErrors) {
	return
}

func (*reportManager) Kind() config.Kind {
	return config.Undefined
}

func (w *reportExecutor) Execute(attrs attribute.Bag, mapper expr.Evaluator) rpc.Status {
	return w.procDispatch(w.ctrs, attrs, mapper, w.hndlr)
}

func (w *reportExecutor) Close() error {
	// Noop: executor does not own the handler, so it cannot close it.
	return nil
}

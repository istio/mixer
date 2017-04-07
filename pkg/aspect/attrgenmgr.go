// Copyright 2017 the Istio Authors.
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
	"github.com/golang/glog"
	rpc "github.com/googleapis/googleapis/google/rpc"

	dpb "istio.io/api/mixer/v1/config/descriptor"
	"istio.io/mixer/pkg/adapter"
	apb "istio.io/mixer/pkg/aspect/config"
	"istio.io/mixer/pkg/attribute"
	"istio.io/mixer/pkg/config"
	"istio.io/mixer/pkg/config/descriptor"
	cpb "istio.io/mixer/pkg/config/proto"
	"istio.io/mixer/pkg/expr"
	"istio.io/mixer/pkg/status"
)

type (
	attrGenMgr  struct{}
	attrGenExec struct {
		aspect adapter.AttributesGenerator
		params *apb.AttributeGeneratorsParams
	}
)

func newAttrGenMgr() PreprocessManager {
	return &attrGenMgr{}
}

func (attrGenMgr) Kind() config.Kind {
	return config.AttributesKind
}

func (attrGenMgr) DefaultConfig() (c config.AspectParams) {
	// NOTE: The default config leads to the generation of no new attributes.
	return &apb.AttributeGeneratorsParams{}
}

func (attrGenMgr) ValidateConfig(c config.AspectParams, v expr.Validator, df descriptor.Finder) (cerrs *adapter.ConfigErrors) {
	params := c.(*apb.AttributeGeneratorsParams)
	attrs := make([]string, 0, len(params.GeneratedAttributes))
	for _, descr := range params.GeneratedAttributes {
		attrs = append(attrs, descr.Name)
	}
	for _, name := range attrs {
		if a := df.GetAttribute(name); a != nil {
			cerrs = cerrs.Appendf(
				"generated_attributes",
				"Attribute '%s' is already configured. It may not be re-generated.",
				name)
		}
	}
	return
}

func (attrGenMgr) NewPreprocessExecutor(cfg *cpb.Combined, b adapter.Builder, env adapter.Env, df descriptor.Finder) (PreprocessExecutor, error) {
	agb := b.(adapter.AttributesGeneratorBuilder)
	ag, err := agb.BuildAttributesGenerator(env, cfg.Builder.Params.(config.AspectParams))
	if err != nil {
		return nil, err
	}
	return &attrGenExec{aspect: ag, params: cfg.Aspect.Params.(*apb.AttributeGeneratorsParams)}, nil
}

func (e *attrGenExec) Execute(attrs attribute.Bag, mapper expr.Evaluator) (*PreprocessResult, rpc.Status) {
	attrGen := e.aspect
	in, err := evalAll(e.params.Labels, attrs, mapper)
	if err != nil {
		errMsg := "Could not evaluate label expressions for attribute generation."
		glog.Error(errMsg, err)
		return nil, status.WithInternal(errMsg)
	}
	out, err := attrGen.Generate(in)
	if err != nil {
		errMsg := "Attribute generation failed."
		glog.Error(errMsg, err)
		return nil, status.WithInternal(errMsg)
	}
	bag := attribute.GetMutableBag(nil)
	adl := attributeDescriptorList(e.params.GeneratedAttributes)
	for key, val := range out {
		if adl.contains(key) {
			// TODO: type validation?
			bag.Set(key, val)
		}
		// TODO: should we return a failure here (produced an attribute
		// that isn't in descriptor list)?
	}
	return &PreprocessResult{Attrs: bag}, status.OK
}

func (e *attrGenExec) Close() error { return e.aspect.Close() }

type attributeDescriptorList []*dpb.AttributeDescriptor

func (l attributeDescriptorList) contains(name string) bool {
	for _, d := range l {
		if d.Name == name {
			return true
		}
	}
	return false
}
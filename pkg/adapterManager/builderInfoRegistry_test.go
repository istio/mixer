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

package adapterManager

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"

	"istio.io/mixer/pkg/adapter"
	"istio.io/mixer/pkg/adapter/config"
	sample_report "istio.io/mixer/pkg/template/sample/report"
)

type TestBuilderInfoInventory struct {
	name string
}

func createBuilderInfo(name string) adapter.BuilderInfo {
	return adapter.BuilderInfo{
		Name:                   name,
		Description:            "mock adapter for testing",
		CreateHandlerBuilderFn: func() config.HandlerBuilder { return fakeHandlerBuilder{} },
		SupportedTemplates:     []adapter.SupportedTemplates{adapter.SampleProcessorTemplate},
		DefaultConfig:          &types.Empty{},
		ValidateConfig:         func(c proto.Message) error { return nil },
	}
}

func (t *TestBuilderInfoInventory) getNewGetBuilderInfoFn() adapter.BuilderInfo {
	return createBuilderInfo(t.name)
}

type fakeHandlerBuilder struct{}

func (fakeHandlerBuilder) ConfigureSample(typeParams map[string]*sample_report.Type) error { return nil }
func (fakeHandlerBuilder) Build(cnfg proto.Message) (config.Handler, error)                { return fakeHandler{}, nil }

type fakeHandler struct{}

func (fakeHandler) Close() error { return nil }
func (fakeHandler) ReportSample(instances []*sample_report.Instance) error {
	return errors.New("not implemented")
}

func TestRegisterSampleProcessor(t *testing.T) {
	var a *sample_report.SampleProcessorBuilder
	fmt.Println(reflect.TypeOf(a).Elem())

	testBuilderInfoInventory := TestBuilderInfoInventory{"foo"}
	reg := newRegistry2([]adapter.GetBuilderInfoFn{testBuilderInfoInventory.getNewGetBuilderInfoFn}, DoesBuilderSupportsTemplate)

	builderInfo, ok := reg.FindBuilderInfo(testBuilderInfoInventory.name)
	if !ok {
		t.Errorf("No builderInfo by name %s, expected %v", testBuilderInfoInventory.name, testBuilderInfoInventory)
	}

	testBuilderInfoObj := testBuilderInfoInventory.getNewGetBuilderInfoFn()
	if testBuilderInfoObj.Name != builderInfo.Name {
		t.Errorf("reg.FindBuilderInfo(%s) expected builderInfo '%v', actual '%v'", testBuilderInfoObj.Name, testBuilderInfoObj, builderInfo)
	}
}

func TestCollisionSameNameAdapter(t *testing.T) {
	testBuilderInfoInventory := TestBuilderInfoInventory{"some name that they both have"}
	testBuilderInfoInventory2 := TestBuilderInfoInventory{"some name that they both have"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected to recover from panic registering duplicate adapter, but recover was nil.")
		}
	}()

	_ = newRegistry2([]adapter.GetBuilderInfoFn{
		testBuilderInfoInventory.getNewGetBuilderInfoFn,
		testBuilderInfoInventory2.getNewGetBuilderInfoFn}, DoesBuilderSupportsTemplate,
	)

	t.Error("Should not reach this statement due to panic.")
}

func TestMissingDefaultValue(t *testing.T) {
	builderCreatorInventory := TestBuilderInfoInventory{"foo"}
	builderInfo := builderCreatorInventory.getNewGetBuilderInfoFn()
	builderInfo.DefaultConfig = nil

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected to recover from panic due to missing DefaultValue in BuilderInfo, " +
				"but recover was nil.")
		}
	}()

	_ = newRegistry2([]adapter.GetBuilderInfoFn{func() adapter.BuilderInfo { return builderInfo }}, DoesBuilderSupportsTemplate)

	t.Error("Should not reach this statement due to panic.")
}

func TestMissingValidateConfigFn(t *testing.T) {
	builderCreatorInventory := TestBuilderInfoInventory{"foo"}
	builderInfo := builderCreatorInventory.getNewGetBuilderInfoFn()
	builderInfo.ValidateConfig = nil

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected to recover from panic due to missing ValidateConfig in BuilderInfo, " +
				"but recover was nil.")
		}
	}()

	_ = newRegistry2([]adapter.GetBuilderInfoFn{func() adapter.BuilderInfo { return builderInfo }}, DoesBuilderSupportsTemplate)

	t.Error("Should not reach this statement due to panic.")
}

func TestHandlerMap(t *testing.T) {
	testBuilderInfoInventory := TestBuilderInfoInventory{"foo"}
	testBuilderInfoInventory2 := TestBuilderInfoInventory{"bar"}

	mp := BuilderInfoMap([]adapter.GetBuilderInfoFn{
		testBuilderInfoInventory.getNewGetBuilderInfoFn,
		testBuilderInfoInventory2.getNewGetBuilderInfoFn,
	}, DoesBuilderSupportsTemplate)

	if _, found := mp["foo"]; !found {
		t.Error("got nil, want foo")
	}
	if _, found := mp["bar"]; !found {
		t.Error("got nil, want bar")
	}
}

type badHandlerBuilder struct{}

func (badHandlerBuilder) DefaultConfig() proto.Message         { return nil }
func (badHandlerBuilder) ValidateConfig(c proto.Message) error { return nil }

// This misspelled function cause the Builder to not implement SampleProcessorBuilder
func (fakeHandlerBuilder) MisspelledXXConfigureSample(typeParams map[string]*sample_report.Type) error {
	return nil
}
func (badHandlerBuilder) Build(cnfg proto.Message) (config.Handler, error) { return fakeHandler{}, nil }

func TestBuilderNotImplementRightTemplateInterface(t *testing.T) {
	badHandlerBuilderBuilderInfo1 := func() adapter.BuilderInfo {
		return adapter.BuilderInfo{
			Name:                   "badAdapter1",
			Description:            "mock adapter for testing",
			CreateHandlerBuilderFn: func() config.HandlerBuilder { return badHandlerBuilder{} },
			SupportedTemplates:     []adapter.SupportedTemplates{adapter.SampleProcessorTemplate},
		}
	}
	badHandlerBuilderBuilderInfo2 := func() adapter.BuilderInfo {
		return adapter.BuilderInfo{
			Name:                   "badAdapter1",
			Description:            "mock adapter for testing",
			CreateHandlerBuilderFn: func() config.HandlerBuilder { return badHandlerBuilder{} },
			SupportedTemplates:     []adapter.SupportedTemplates{adapter.SampleProcessorTemplate},
		}
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected to recover from panic registering bad builder that does not implement Builders " +
				"for all supported templates, but recover was nil.")
		}
	}()

	_ = newRegistry2([]adapter.GetBuilderInfoFn{
		badHandlerBuilderBuilderInfo1, badHandlerBuilderBuilderInfo2}, DoesBuilderSupportsTemplate,
	)

	t.Error("Should not reach this statement due to panic.")
}

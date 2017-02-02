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

package statsd

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/status"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/duration"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/cactus/go-statsd-client/statsd/statsdtest"
	"istio.io/mixer/adapter/statsd/config"
	"istio.io/mixer/pkg/adapter"
	"istio.io/mixer/pkg/adapter/test"
)

func TestInvariants(t *testing.T) {
	test.AdapterInvariants(Register, t)
}

func TestNewBuilder(t *testing.T) {
	conf = &config.Params{
		Address:                   "localhost:8125",
		Prefix:                    "",
		FlushInterval:             &duration.Duration{Seconds: 0, Nanos: int32(300 * time.Millisecond)},
		FlushBytes:                512,
		SamplingRate:              1.0,
		MetricNameTemplateStrings: map[string]string{"a": `{{ .apiMethod "-" .responseCode }}`},
	}
	b := newBuilder()

	if err := b.Close(); err != nil {
		t.Errorf("b.Close() = %s, expected no err", err)
	}
}

func TestNewBuilder_BadTemplate(t *testing.T) {
	conf.MetricNameTemplateStrings = map[string]string{"badtemplate": `{{if 1}}`}
	defer func() {
		if r := recover(); r == nil {
			t.Error("newBuilder() didn't panic")
		}
	}()
	_ = newBuilder()
	t.Fail()
}

func TestNewBuilder_BadStatsdConfig(t *testing.T) {
	conf.Address = "notaurl:notaport"
	defer func() {
		if r := recover(); r == nil {
			t.Error("newBuilder() didn't panic")
		}
	}()
	_ = newBuilder()
	t.Fail()
}

func TestValidateConfig(t *testing.T) {
	cases := []struct {
		conf      proto.Message
		errString string
	}{
		{&config.Params{}, ""},
		{&config.Params{MetricNameTemplateStrings: map[string]string{"a": `{{ .apiMethod "-" .responseCode }}`}}, ""},
		{&config.Params{MetricNameTemplateStrings: map[string]string{"badtemplate": `{{if 1}}`}}, "badtemplate"},
	}
	for idx, c := range cases {
		b := &builder{}
		errString := ""
		if err := b.ValidateConfig(c.conf); err != nil {
			errString = err.Error()
		}
		if !strings.Contains(errString, c.errString) {
			t.Errorf("[%d] b.ValidateConfig(c.conf) = '%s'; want errString containing '%s'", idx, errString, c.errString)
		}
	}
}

func TestValidateConfig_PanicsWithWrongConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("b.ValidateConfig(<wrong type>) didn't panic")
		}
	}()
	b := &builder{}
	if err := b.ValidateConfig(&status.Status{}); err != nil {
		t.Errorf("b.ValidateConfig(&badconfig.Params{}) = %v, wanted panic", err)
	}
	t.Fail()
}

func TestNewMetric(t *testing.T) {
	conf = &config.Params{
		Address:                   "localhost:8125",
		Prefix:                    "",
		FlushInterval:             &duration.Duration{Seconds: 0, Nanos: int32(300 * time.Millisecond)},
		FlushBytes:                512,
		SamplingRate:              1.0,
		MetricNameTemplateStrings: map[string]string{"a": `{{(.apiMethod) "-" (.responseCode)}}`},
	}
	b := newBuilder()
	if _, err := b.NewMetricsAspect(test.NewEnv(t), &status.Status{}, nil); err == nil {
		t.Error("b.NewMetrics(test.NewEnv(t), &status.Status{}) = _, nil; wanted err for wrong config type")
	}

	masp, err := b.NewMetricsAspect(test.NewEnv(t), &config.Params{}, nil)
	if err != nil {
		t.Errorf("b.NewMetrics(test.NewEnv(t), &config.Params{}) = %s, wanted no err", err)
	}
	asp := masp.(*aspect)
	if asp.client != b.client {
		t.Errorf("asp.client = %v, wanted b.client (%v) to verify shared connection", asp.client, b.client)
	}
}

func TestRecord(t *testing.T) {
	var templateMetricName = "methodCode"

	conf = &config.Params{
		Address:       "localhost:8125",
		Prefix:        "",
		FlushInterval: &duration.Duration{Seconds: 0, Nanos: int32(300 * time.Millisecond)},
		FlushBytes:    512,
		SamplingRate:  1.0,
		MetricNameTemplateStrings: map[string]string{
			templateMetricName: `{{.apiMethod}}-{{.responseCode}}`,
			"invalidTemplate":  `{{ .apiMethod "-" .responseCode }}`, // fails at execute time, not template parsing time
		},
	}

	validGauge := adapter.Value{
		Name:        "foo",
		Kind:        adapter.Gauge,
		Labels:      make(map[string]interface{}),
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		MetricValue: int64(123),
	}
	invalidGauge := validGauge
	invalidGauge.MetricValue = "bar"

	validCounter := adapter.Value{
		Name:        "bar",
		Kind:        adapter.Counter,
		Labels:      make(map[string]interface{}),
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		MetricValue: int64(123),
	}
	invalidCounter := validCounter
	invalidCounter.MetricValue = 1.0

	invalidKind := validCounter
	invalidKind.Kind = adapter.Kind(37)

	methodCodeMetric := validCounter
	methodCodeMetric.Name = templateMetricName // this needs to match the name in conf.MetricNameTemplateStrings
	methodCodeMetric.Labels["apiMethod"] = "methodName"
	methodCodeMetric.Labels["responseCode"] = "500"
	expectedMetricName := methodCodeMetric.Labels["apiMethod"].(string) + "-" + methodCodeMetric.Labels["responseCode"].(string)

	invalidTemplateMetric := validGauge
	invalidTemplateMetric.Name = "invalidTemplate" // this needs to match the name in conf.MetricNameTemplateStrings
	invalidTemplateMetric.Labels["apiMethod"] = "some method"
	invalidTemplateMetric.Labels["responseCode"] = "3"

	cases := []struct {
		vals      []adapter.Value
		errString string
	}{
		{[]adapter.Value{}, ""},
		{[]adapter.Value{validGauge}, ""},
		{[]adapter.Value{validCounter}, ""},
		{[]adapter.Value{methodCodeMetric}, ""},
		{[]adapter.Value{validCounter, validGauge}, ""},
		{[]adapter.Value{validCounter, validGauge, methodCodeMetric}, ""},
		{[]adapter.Value{invalidKind}, "unknown metric kind"},
		{[]adapter.Value{invalidCounter}, "could not record"},
		{[]adapter.Value{invalidGauge}, "could not record"},
		{[]adapter.Value{invalidTemplateMetric}, "failed to create metric name"},
		{[]adapter.Value{validGauge, invalidGauge}, "could not record"},
		{[]adapter.Value{methodCodeMetric, invalidCounter}, "could not record"},
	}
	for idx, c := range cases {
		b := newBuilder()
		rs := statsdtest.NewRecordingSender()
		cl, err := statsd.NewClientWithSender(rs, "")
		if err != nil {
			t.Errorf("statsd.NewClientWithSender(rs, \"\") = %s; wanted no err", err)
		}
		b.client = cl

		m, err := b.NewMetricsAspect(test.NewEnv(t), conf, nil)
		if err != nil {
			t.Errorf("newBuilder().NewMetrics(test.NewEnv(t), conf) = _, %s; wanted no err", err)
		}
		if err := m.Record(c.vals); err != nil {
			if c.errString == "" {
				t.Errorf("[%d] m.Record(c.vals) = %s; wanted no err", idx, err)
			}
			if !strings.Contains(err.Error(), c.errString) {
				t.Errorf("[%d] m.Record(c.vals) = %s; wanted err containing %s", idx, err.Error(), c.errString)
			}
		}
		if err := m.Close(); err != nil {
			t.Errorf("m.Close() = %s; wanted no err", err)
		}
		if c.errString != "" {
			continue
		}

		metrics := rs.GetSent()
		for _, val := range c.vals {
			name := val.Name
			if val.Name == templateMetricName {
				name = expectedMetricName
			}
			m := metrics.CollectNamed(name)
			if len(m) < 1 {
				t.Errorf("[%d] metrics.CollectNamed(%s) returned no stats, expected one", idx, name)
			}
		}
	}
}

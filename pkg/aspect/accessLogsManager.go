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

package aspect

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"
	"time"

	rpc "github.com/googleapis/googleapis/google/rpc"

	"istio.io/mixer/pkg/adapter"
	aconfig "istio.io/mixer/pkg/aspect/config"
	"istio.io/mixer/pkg/attribute"
	"istio.io/mixer/pkg/config"
	"istio.io/mixer/pkg/expr"
)

type (
	accessLogsManager struct{}

	accessLogsWrapper struct {
		name     string
		aspect   adapter.AccessLogsAspect
		labels   map[string]string // label name -> expression
		template *template.Template
	}
)

const (
	// TODO: revisit when well-known attributes are defined.
	commonLogFormat = `{{or (.originIp) "-"}} - {{or (.source_user) "-"}} ` +
		`[{{or (.timestamp.Format "02/Jan/2006:15:04:05 -0700") "-"}}] "{{or (.method) "-"}} ` +
		`{{or (.url) "-"}} {{or (.protocol) "-"}}" {{or (.responseCode) "-"}} {{or (.responseSize) "-"}}`
	// TODO: revisit when well-known attributes are defined.
	combinedLogFormat = commonLogFormat + ` "{{or (.referer) "-"}}" "{{or (.user_agent) "-"}}"`
)

var (
	commonLogLabels = map[string]string{
		"originIp":     "origin.ip",
		"source_user":  "origin.user",
		"timestamp":    "request.time",
		"method":       `request.method | ""`,
		"url":          "request.path",
		"protocol":     "request.scheme",
		"responseCode": "response.code",
		"responseSize": "response.size",
	}

	// TODO: revisit when well-known attributes are defined
	combinedLogLabels = map[string]string{
		"originIp":     "origin.ip",
		"source_user":  "origin.user",
		"timestamp":    "request.time",
		"method":       `request.method | ""`,
		"url":          "request.path",
		"protocol":     "request.scheme",
		"responseCode": "response.code",
		"responseSize": "response.size",
		"referer":      "request.referer",
		"user_agent":   "request.user-agent",
	}

	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// NewAccessLogsManager returns a manager for the access logs aspect.
func NewAccessLogsManager() Manager {
	return accessLogsManager{}
}

func (m accessLogsManager) NewAspect(c *config.Combined, a adapter.Builder, env adapter.Env) (Wrapper, error) {
	logCfg := c.Aspect.Params.(*aconfig.AccessLogsParams)

	var labels map[string]string
	var templateStr string
	switch logCfg.Log.LogFormat {
	case aconfig.COMMON:
		templateStr = commonLogFormat
		labels = commonLogLabels
	case aconfig.COMBINED:
		templateStr = combinedLogFormat
		labels = combinedLogLabels
	case aconfig.CUSTOM:
		fallthrough
	default:
		// Hack because user's can't give us descriptors yet. For now custom template can be created by
		// defining a "template" input. This is not documented anywhere but here.
		templateStr = c.Aspect.Inputs["template"]
		labels = logCfg.Log.Labels
	}

	// TODO: when users can provide us with descriptors, this error can be removed due to validation
	tmpl, err := template.New("accessLogsTemplate").Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("log %s failed to parse template '%s' with err: %s", logCfg.LogName, templateStr, err)
	}

	asp, err := a.(adapter.AccessLogsBuilder).NewAccessLogsAspect(env, c.Builder.Params.(adapter.AspectConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create aspect for log %s with err: %s", logCfg.LogName, err)
	}

	return &accessLogsWrapper{
		logCfg.LogName,
		asp,
		labels,
		tmpl,
	}, nil
}

func (accessLogsManager) Kind() Kind { return AccessLogsKind }
func (accessLogsManager) DefaultConfig() adapter.AspectConfig {
	return &aconfig.AccessLogsParams{
		LogName: "access_log",
		Log: &aconfig.AccessLogsParams_AccessLog{
			LogFormat: aconfig.COMMON,
		},
	}
}

func (accessLogsManager) ValidateConfig(c adapter.AspectConfig) (ce *adapter.ConfigErrors) {
	cfg := c.(*aconfig.AccessLogsParams)
	if cfg.Log == nil {
		ce = ce.Appendf("Log", "An AccessLog entry must be provided.")
		return
	}
	if cfg.Log.LogFormat != aconfig.CUSTOM {
		// If it's not custom we're using our own configs, so we're fine.
		return nil
	}
	// TODO: validate custom templates when users can provide us with descriptors
	return nil
}

func (e *accessLogsWrapper) Close() error {
	return e.aspect.Close()
}

func (e *accessLogsWrapper) Execute(attrs attribute.Bag, mapper expr.Evaluator, ma APIMethodArgs) (*Output, error) {
	labels, err := evalAll(e.labels, attrs, mapper)
	if err != nil {
		return &Output{Code: rpc.INTERNAL}, fmt.Errorf("failed to eval labels for log %s with err: %s", e.name, err)
	}

	// TODO: better way to ensure timestamp is available if not supplied
	// in Report() requests.
	if _, found := labels["timestamp"]; !found {
		labels["timestamp"] = time.Now()
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	if err := e.template.Execute(buf, labels); err != nil {
		return &Output{Code: rpc.INTERNAL}, err
	}
	payload := buf.String()
	buf.Reset()
	bufferPool.Put(buf)

	entry := adapter.LogEntry{
		LogName:     e.name,
		Labels:      labels,
		TextPayload: payload,
	}
	if err := e.aspect.LogAccess([]adapter.LogEntry{entry}); err != nil {
		return nil, fmt.Errorf("failed to log %s with err: %s", e.name, err)
	}
	return &Output{Code: rpc.OK}, nil
}

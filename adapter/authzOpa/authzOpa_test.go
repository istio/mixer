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

package authzOpa

// NOTE: This test will eventually be auto-generated so that it automatically supports all CHECK and QUOTA
//       templates known to Mixer. For now, it's manually curated.

import (
	"context"
	"testing"

	rpc "github.com/googleapis/googleapis/google/rpc"

	"istio.io/mixer/adapter/authzOpa/config"
	"istio.io/mixer/pkg/adapter/test"
	"istio.io/mixer/template/authz"
)

var (
	policy = `package mixerauthz
	    policy = [
	      {
	        "rule": {
	          "verbs": [
	            "storage.buckets.get"
	          ],
	          "users": [
	            "bucket-admins"
	          ]
	        }
	      }
	    ]
	
	    default allow = false
	
	    allow = true {
	      rule = policy[_].rule
	      input.user = rule.users[_]
	      input.verb = rule.verbs[_]
	    }`
	checkMethod = "data.mixerauthz.allow"
	cfg         = &config.Params{
		Policy:      policy,
		CheckMethod: checkMethod,
	}
)

func TestAccepted(t *testing.T) {
	info := GetInfo()

	if !contains(info.SupportedTemplates, authz.TemplateName) {
		t.Error("Didn't find all expected supported templates")
	}

	b := info.NewBuilder().(*builder)
	b.SetAdapterConfig(cfg)
	if err := b.Validate(); err != nil {
		t.Errorf("Got error %v, expecting success", err)
	}

	handler, err := b.Build(context.Background(), test.NewEnv(t))
	if err != nil {
		t.Errorf("Got error %v, expecting success", err)
	}

	authzHandler := handler.(authz.Handler)
	instance := authz.Instance{
		Principal: make(map[string]interface {
		}),
		Resource: make(map[string]interface {
		}),
		Verb: "storage.buckets.get",
	}
	instance.Principal["user"] = "bucket-admins"

	result, err := authzHandler.HandleAuthz(context.Background(), &instance)
	if err != nil {
		t.Errorf("Got error %v, expecting success", err)
	}

	if result.Status.Code != int32(rpc.OK) {
		t.Errorf("Got error %v, expecting success", err)
	}
}

func TestRejectedRequest(t *testing.T) {
	info := GetInfo()

	if !contains(info.SupportedTemplates, authz.TemplateName) {
		t.Error("Didn't find all expected supported templates")
	}

	b := info.NewBuilder().(*builder)
	b.SetAdapterConfig(cfg)
	if err := b.Validate(); err != nil {
		t.Errorf("Got error %v, expecting success", err)
	}

	handler, err := b.Build(context.Background(), test.NewEnv(t))
	if err != nil {
		t.Errorf("Got error %v, expecting success", err)
	}

	authzHandler := handler.(authz.Handler)
	instance := authz.Instance{
		Principal: make(map[string]interface {
		}),
		Resource: make(map[string]interface {
		}),
		Verb: "storage.buckets.put",
	}
	instance.Principal["user"] = "bucket-admins"

	result, err := authzHandler.HandleAuthz(context.Background(), &instance)
	if err != nil {
		t.Errorf("Got error %v, expecting success", err)
	}

	if result.Status.Code != int32(rpc.PERMISSION_DENIED) {
		t.Errorf("Got error %v, expecting success", err)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

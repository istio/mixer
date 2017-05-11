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
package config

import (
	"reflect"
	"testing"
)

type KVMgr struct {
	Store   KeyValueStore
	cleanup func()
}

func (k *KVMgr) Cleanup() {
	k.Store.Close()
	k.cleanup()
}

func testStore(t *testing.T, kvMgrfn func() *KVMgr) {
	GOODKEYS := []string{
		"/scopes/global/adapters",
		"/scopes/global/descriptors",
		"/scopes/global/subjects/global/rules",
		"/scopes/global/subjects/svc1.ns.cluster.local/rules",
	}

	table := []struct {
		desc       string
		keys       []string
		listPrefix string
		listKeys   []string
	}{
		{"goodkeys", GOODKEYS, "/scopes/global/subjects",
			[]string{"/scopes/global/subjects/global/rules",
				"/scopes/global/subjects/svc1.ns.cluster.local/rules"},
		},
		{"goodkeys", GOODKEYS, "/scopes/", GOODKEYS},
	}

	for _, tt := range table {
		km := kvMgrfn()
		s := km.Store
		t.Run(tt.desc, func(t1 *testing.T) {
			var found bool
			badkey := "a/b"
			_, _, found = s.Get(badkey)
			if found {
				t.Errorf("Unexpectedly found %s", badkey)
			}
			var val string
			// create keys
			for _, key := range tt.keys {
				kc := key
				_, err := s.Set(key, kc)
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", key, err)
				}
				val, _, found = s.Get(key)
				if !found || kc != val {
					t.Errorf("got %s\nwant %s", val, kc)
				}
			}
			k, _, err := s.List(tt.listPrefix, true)
			if err != nil {
				t.Error("Unexpected error", err)
			}
			if !reflect.DeepEqual(k, tt.listKeys) {
				t.Errorf("Got %s\nWant %s\n", k, tt.listKeys)
			}
			err = s.Delete(k[1])
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			_, _, found = s.Get(k[1])
			if found {
				t.Errorf("Unexpectedly found %s", k[1])
			}

		})
		s.Close()
		km.Cleanup()
	}
}

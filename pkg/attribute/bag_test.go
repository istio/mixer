// Copyright 2016 Istio Authors
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

package attribute

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	mixerpb "istio.io/api/mixer/v1"
)

var (
	t9  = time.Date(2001, 1, 1, 1, 1, 1, 9, time.UTC)
	t10 = time.Date(2001, 1, 1, 1, 1, 1, 10, time.UTC)
	t42 = time.Date(2001, 1, 1, 1, 1, 1, 42, time.UTC)

	d1 = 42 * time.Second
	d2 = 34 * time.Second
)

func TestBag(t *testing.T) {
	sm1 := mixerpb.StringMap{Entries: map[int32]int32{-16: -16}}
	sm2 := mixerpb.StringMap{Entries: map[int32]int32{-17: -17}}
	m1 := map[string]string{"N16": "N16"}
	m3 := map[string]string{"N42": "FourtyTwo"}

	attrs := mixerpb.Attributes{
		Words:      []string{"N1", "N2", "N3", "N4", "N5", "N6", "N7", "N8", "N9", "N10", "N11", "N12", "N13", "N14", "N15", "N16", "N17"},
		Strings:    map[int32]int32{-1: -1, -2: -2},
		Int64S:     map[int32]int64{-3: 3, -4: 4},
		Doubles:    map[int32]float64{-5: 5.0, -6: 6.0},
		Bools:      map[int32]bool{-7: true, -8: false},
		Timestamps: map[int32]time.Time{-9: t9, -10: t10},
		Durations:  map[int32]time.Duration{-11: d1},
		Bytes:      map[int32][]uint8{-12: {12}, -13: {13}},
		StringMaps: map[int32]mixerpb.StringMap{-14: sm1, -15: sm2},
	}

	ab, err := GetBagFromProto(&attrs, nil)
	if err != nil {
		t.Errorf("Unable to start request: %v", err)
	}

	// override a bunch of values
	ab.Set("N2", "42")
	ab.Set("N4", int64(42))
	ab.Set("N6", float64(42.0))
	ab.Set("N8", true)
	ab.Set("N10", t42)
	ab.Set("N11", d2)
	ab.Set("N13", []byte{42})
	ab.Set("N15", m3)

	// make sure the overrides worked and didn't disturb non-overridden values
	results := []struct {
		name  string
		value interface{}
	}{
		{"N1", "N1"},
		{"N2", "42"},
		{"N3", int64(3)},
		{"N4", int64(42)},
		{"N5", 5.0},
		{"N6", 42.0},
		{"N7", true},
		{"N8", true},
		{"N9", t9},
		{"N10", t42},
		{"N11", d2},
		{"N12", []byte{12}},
		{"N13", []byte{42}},
		{"N14", m1},
		{"N15", m3},
	}

	for _, r := range results {
		t.Run(r.name, func(t *testing.T) {
			v, found := ab.Get(r.name)
			if !found {
				t.Error("Got false, expecting true")
			}

			if !reflect.DeepEqual(v, r.value) {
				t.Errorf("Got %v, expected %v", v, r.value)
			}
		})
	}

	if _, found := ab.Get("XYZ"); found {
		t.Error("XYZ was found")
	}

	// try another level of overrides just to make sure that path is OK
	child := ab.Child()
	child.Set("N2", "31415692")
	r, found := ab.Get("N2")
	if !found || r.(string) != "42" {
		t.Error("N2 has wrong value")
	}
}

/*
func TestStringMapEdgeCase(t *testing.T) {
	// ensure coverage for some obscure logging paths

	d := dictionary{1: "N1", 2: "N2"}
	rb := GetMutableBag(nil)
	attrs := &mixerpb.Attributes{}

	// empty to non-empty
	sm1 := mixerpb.StringMap{Map: map[int32]string{2: "Two"}}
	attrs.StringMapAttributes = map[int32]mixerpb.StringMap{1: sm1}
	_ = rb.update(d, attrs)

	// non-empty to non-empty
	sm1 = mixerpb.StringMap{Map: map[int32]string{}}
	attrs.StringMapAttributes = map[int32]mixerpb.StringMap{1: sm1, 2: sm1}
	_ = rb.update(d, attrs)

	// non-empty to empty
	attrs.DeletedAttributes = []int32{1}
	attrs.StringMapAttributes = map[int32]mixerpb.StringMap{}
	_ = rb.update(d, attrs)
}

func TestBadStringMapKey(t *testing.T) {
	// ensure we handle bogus on-the-wire string map key indices

	sm1 := mixerpb.StringMap{Map: map[int32]string{16: "Sixteen"}}

	attr := mixerpb.Attributes{
		Dictionary:          dictionary{1: "N1"},
		StringMapAttributes: map[int32]mixerpb.StringMap{1: sm1},
	}

	am := NewManager()
	at := am.NewTracker()
	defer at.Done()

	_, err := at.ApplyProto(&attr)
	if err == nil {
		t.Error("Successfully updated attributes, expected an error")
	}
}
*/

func TestMerge(t *testing.T) {
	mb := GetMutableBag(empty)

	c1 := mb.Child()
	c2 := mb.Child()

	c1.Set("STRING1", "A")
	c2.Set("STRING2", "B")

	if err := mb.Merge(c1, c2); err != nil {
		t.Errorf("Got %v, expecting success", err)
	}

	if v, ok := mb.Get("STRING1"); !ok || v.(string) != "A" {
		t.Errorf("Got %v, expected A", v)
	}

	if v, ok := mb.Get("STRING2"); !ok || v.(string) != "B" {
		t.Errorf("Got %v, expected B", v)
	}
}

func TestMergeErrors(t *testing.T) {
	mb := GetMutableBag(empty)

	c1 := mb.Child()
	c2 := mb.Child()

	c1.Set("FOO", "X")
	c2.Set("FOO", "Y")

	if err := mb.Merge(c1, c2); err == nil {
		t.Error("Got success, expected failure")
	} else if !strings.Contains(err.Error(), "FOO") {
		t.Errorf("Expected error to contain the word FOO, got %s", err.Error())
	}
}

func TestEmpty(t *testing.T) {
	b := &emptyBag{}

	if names := b.Names(); len(names) > 0 {
		t.Errorf("Get len %d, expected 0", len(names))
	}

	if _, ok := b.Get("XYZ"); ok {
		t.Errorf("Got true, expected false")
	}

	b.Done()
}

func TestEmptyRoundTrip(t *testing.T) {
	attrs0 := mixerpb.Attributes{}
	attrs1 := mixerpb.Attributes{}
	mb := GetMutableBag(nil)
	mb.ToProto(&attrs1, nil)

	if !reflect.DeepEqual(attrs0, attrs1) {
		t.Error("Expecting equal attributes, got a delta")
	}
}

func TestProtoBag(t *testing.T) {
	globalDict := []string{"G0", "G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9"}
	messageDict := []string{"M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9", "M10"}

	revGlobalDict := make(map[string]int32)
	for k, v := range globalDict {
		revGlobalDict[v] = int32(k)
	}

	sm := mixerpb.StringMap{Entries: map[int32]int32{-6: -7}}

	attrs := mixerpb.Attributes{
		Words:      messageDict,
		Strings:    map[int32]int32{4: 5},
		Int64S:     map[int32]int64{6: 42},
		Doubles:    map[int32]float64{7: 42.0},
		Bools:      map[int32]bool{-1: true},
		Timestamps: map[int32]time.Time{-2: t9},
		Durations:  map[int32]time.Duration{-3: d1},
		Bytes:      map[int32][]uint8{-4: {11}},
		StringMaps: map[int32]mixerpb.StringMap{-5: sm},
	}

	cases := []struct {
		name  string
		value interface{}
	}{
		{"G4", "G5"},
		{"G6", int64(42)},
		{"G7", 42.0},
		{"M1", true},
		{"M2", t9},
		{"M3", d1},
		{"M4", []byte{11}},
		{"M5", map[string]string{"M6": "M7"}},
	}

	for i := 0; i < 2; i++ {
		pb, err := GetBagFromProto(&attrs, globalDict)
		if err != nil {
			t.Fatalf("GetBagFromProto failed with %v", err)
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				v, ok := pb.Get(c.name)
				if !ok {
					t.Error("Got false, expected true")
				}

				if ok, _ := compareAttributeValues(v, c.value); !ok {
					t.Errorf("Got %v, expected %v", v, c.value)
				}
			})
		}

		// make sure all the expected names are there
		names := pb.Names()
		for _, cs := range cases {
			found := false
			for _, n := range names {
				if cs.name == n {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Could not find attribute name %s", cs.name)
			}
		}

		// try out round-tripping
		mb := GetMutableBag(pb)
		for _, n := range names {
			v, _ := pb.Get(n)
			mb.Set(n, v)
		}

		mb.ToProto(&attrs, revGlobalDict)

		pb.Done()
	}
}

func TestProtoBag_Errors(t *testing.T) {
	globalDict := []string{"G0", "G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9"}
	messageDict := []string{"M0", "M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9"}

	attrs := mixerpb.Attributes{
		Words:   messageDict,
		Strings: map[int32]int32{-24: 25},
	}

	pb, err := GetBagFromProto(&attrs, globalDict)
	if err == nil {
		t.Error("GetBagFromProto succeeded, expected failure")
	}

	if pb != nil {
		t.Error("GetBagFromProto returned valid bag, expected nil")
	}
}

func init() {
	// bump up the log level so log-only logic runs during the tests, for correctness and coverage.
	_ = flag.Lookup("v").Value.Set("99")
}

func TestMutableBag_Child(t *testing.T) {
	mb := GetMutableBag(nil)
	c1 := mb.Child()
	c2 := mb.Child()
	c3 := mb.Child()
	c31 := c3.Child()
	mb.Done()

	if mb.parent == nil {
		t.Errorf("Unexpectedly freed bag with children %#v", mb)
	}

	if c1.parent != mb {
		t.Errorf("not the correct parent. got %#v\nwant %#v", c1.parent, mb)
	}
	c1.Done()
	if c1.parent != nil {
		t.Errorf("did not free bag c1 %#v", c1)
	}
	c2.Done()
	c3.Done()
	c31.Done()

	mb.parent = nil
	err := withPanic(func() { mb.Child() })

	if err == nil {
		t.Errorf("want panic, got %#v", err)
	}
}

func withPanic(f func()) (ret interface{}) {
	defer func() {
		ret = recover()
	}()

	f()
	return ret
}

func compareAttributeValues(v1, v2 interface{}) (bool, error) {
	var result bool
	switch t1 := v1.(type) {
	case string:
		t2, ok := v2.(string)
		result = ok && t1 == t2
	case int64:
		t2, ok := v2.(int64)
		result = ok && t1 == t2
	case float64:
		t2, ok := v2.(float64)
		result = ok && t1 == t2
	case bool:
		t2, ok := v2.(bool)
		result = ok && t1 == t2
	case time.Time:
		t2, ok := v2.(time.Time)
		result = ok && t1 == t2
	case time.Duration:
		t2, ok := v2.(time.Duration)
		result = ok && t1 == t2

	case []byte:
		t2, ok := v2.([]byte)
		if result = ok && len(t1) == len(t2); result {
			for i := 0; i < len(t1); i++ {
				if t1[i] != t2[i] {
					result = false
					break
				}
			}
		}

	case map[string]string:
		t2, ok := v2.(map[string]string)
		if result = ok && len(t1) == len(t2); result {
			for k, v := range t1 {
				if v != t2[k] {
					result = false
					break
				}
			}
		}

	default:
		return false, fmt.Errorf("unsupported attribute value type: %T", v1)
	}

	return result, nil
}

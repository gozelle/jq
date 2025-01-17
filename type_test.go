package jq_test

import (
	"fmt"
	"github.com/gozelle/jq"
	"math"
	"math/big"
	"testing"
)

func TestTypeOf(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected string
	}{
		{nil, "null"},
		{false, "boolean"},
		{true, "boolean"},
		{0, "number"},
		{3.14, "number"},
		{math.NaN(), "number"},
		{math.Inf(1), "number"},
		{math.Inf(-1), "number"},
		{big.NewInt(10), "number"},
		{"string", "string"},
		{[]interface{}{}, "array"},
		{map[string]interface{}{}, "object"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.value), func(t *testing.T) {
			got := jq.TypeOf(tc.value)
			if got != tc.expected {
				t.Errorf("TypeOf(%v): got %s, expected %s", tc.value, got, tc.expected)
			}
		})
	}
	func() {
		v := []int{0}
		defer func() {
			if got, expected := recover(), "invalid type: []int ([0])"; got != expected {
				t.Errorf("TypeOf(%v) should panic: got %v, expected %v", v, got, expected)
			}
		}()
		_ = jq.TypeOf(v)
	}()
}

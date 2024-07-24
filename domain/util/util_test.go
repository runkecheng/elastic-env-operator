package util

import (
	"strings"
	"testing"
	
	"gotest.tools/assert"
)

func TestMergeStringMap(t *testing.T) {
	assert.Equal(t, MergeStringMap(nil, map[string]string{"a": "1"})["a"], "1")
}

func TestMergeStringMap_1(t *testing.T) {
	assert.Equal(t, MergeStringMap(map[string]string{"a": "1"}, nil)["a"], "1")
}

func TestMergeStringMap_Filter(t *testing.T) {
	assert.Equal(t, len(MergeLabelsWithFilter(map[string]string{"app": "1", "rollouts.kruise.io/stable-revision": "55cc48bb98"}, map[string]string{"a": "2"}, func(key string) bool {
		if strings.HasPrefix(key, "rollouts.kruise.io/") {
			return false
		}
		if strings.HasPrefix(key, "k8slens-") {
			return false
		}
		return true
	})), 2)
}

// SPDX-License-Identifier: MIT

package diff

import (
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

func CreateTwoWayMergePatch(orig, patched, dataStruct any) ([]byte, bool, error) {
	origBytes, err := json.Marshal(orig)
	if err != nil {
		return nil, false, err
	}
	newBytes, err := json.Marshal(patched)
	if err != nil {
		return nil, false, err
	}
	patch, err := strategicpatch.CreateTwoWayMergePatch(origBytes, newBytes, dataStruct)
	if err != nil {
		return nil, false, err
	}
	return patch, string(patch) != "{}", nil
}

func StrategicMergePatch(orig any, patchBytes []byte, dataStruct any) ([]byte, error) {
	origBytes, err := json.Marshal(orig)
	if err != nil {
		return nil, err
	}

	newBytes, err := strategicpatch.StrategicMergePatch(origBytes, patchBytes, dataStruct)
	if err != nil {
		return nil, err
	}
	return newBytes, nil
}

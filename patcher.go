package main

import (
	"encoding/json"
)

func patchBuilder(labels map[string]string) ([]byte, error) {
	patches := []map[string]interface{}{}

	// Ensure the labels field exists
	patches = append(patches, map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels",
		"value": map[string]string{},
	})

	for key, value := range labels {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + key,
			"value": value,
		})
	}

	return json.Marshal(patches)
}

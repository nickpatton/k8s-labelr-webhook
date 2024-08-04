package main

import (
	"encoding/json"
)

func patchBuilder(k8sObject GenericK8sObject) ([]byte, error) {

	labels := map[string]string{
		"sea":  "turtle",
		"land": "mongoose",
		"air":  "falcon",
	}

	annotations := map[string]string{
		"address": "P. Sherman 42 Wallaby Way, Sydney, Austrailia",
	}

	patches := []map[string]interface{}{}

	// If the object doesn't have labels or annotations, we need to add the fields
	// Otherwise, we don't want to squash or stomp th existing labels or annotations
	if k8sObject.ObjectMeta.Labels == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels",
			"value": map[string]string{},
		})
	}

	if k8sObject.ObjectMeta.Annotations == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/annotations",
			"value": map[string]string{},
		})
	}

	logger.Print("assembling the metadata patch with supplied labels and/or annotations")

	for key, value := range labels {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + key,
			"value": value,
		})
	}

	for key, value := range annotations {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/annotations/" + key,
			"value": value,
		})
	}

	// If there's a spec.template.metadata.labels field, we're likely dealing with a deployment, rollout, statefulset, etc.
	// These objects need labels under the spec.template stanza so the pods inherit the labels as well.
	if k8sObject.Spec.Template.Metadata.Labels != nil {
		logger.Print("assembling additional patches for spec.template.metadata fields")

		for key, value := range labels {
			patches = append(patches, map[string]interface{}{
				"op":    "add",
				"path":  "/spec/template/metadata/labels/" + key,
				"value": value,
			})
			patches = append(patches, map[string]interface{}{
				"op":    "add",
				"path":  "/spec/selector/matchLabels/" + key,
				"value": value,
			})
		}

		if k8sObject.Spec.Template.Metadata.Annotations == nil {
			patches = append(patches, map[string]interface{}{
				"op":    "add",
				"path":  "/spec/template/metadata/annotations",
				"value": map[string]string{},
			})
		}
		for key, value := range annotations {
			patches = append(patches, map[string]interface{}{
				"op":    "add",
				"path":  "/spec/template/metadata/annotations/" + key,
				"value": value,
			})
		}
	}

	return json.Marshal(patches)
}

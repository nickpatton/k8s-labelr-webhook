package main

import (
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Creating a struct for unmarshalling object in AdmissionReview into a "generic" K8s object
type GenericK8sObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GenericK8sObjectSpec `json:"spec,omitempty"`
}

type GenericK8sObjectSpec struct {
	Template GenericK8sObjectTemplate `json:"template,omitempty"`
}

type GenericK8sObjectTemplate struct {
	Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
}

func passingAdmissionReviewResponse(admissionReviewRequest *admissionv1.AdmissionReview) admissionv1.AdmissionReview {
	admissionResponse := &admissionv1.AdmissionResponse{}
	admissionResponse.Allowed = true

	var admissionReviewResponse admissionv1.AdmissionReview
	admissionReviewResponse.Response = admissionResponse
	admissionReviewResponse.SetGroupVersionKind(admissionReviewRequest.GroupVersionKind())
	admissionReviewResponse.Response.UID = admissionReviewRequest.Request.UID

	return admissionReviewResponse
}

func patchingAdmissionReviewResponse(admissionReviewRequest *admissionv1.AdmissionReview, patch []byte) admissionv1.AdmissionReview {
	admissionResponse := &admissionv1.AdmissionResponse{}
	admissionResponse.Allowed = true
	admissionResponse.Patch = patch
	patchType := admissionv1.PatchTypeJSONPatch
	admissionResponse.PatchType = &patchType

	var admissionReviewResponse admissionv1.AdmissionReview
	admissionReviewResponse.Response = admissionResponse
	admissionReviewResponse.SetGroupVersionKind(admissionReviewRequest.GroupVersionKind())
	admissionReviewResponse.Response.UID = admissionReviewRequest.Request.UID

	return admissionReviewResponse
}

package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	codecs = serializer.NewCodecFactory(runtime.NewScheme())
	logger = log.New(os.Stdout, "http: ", log.LstdFlags)
)

func main() {
	cert, err := tls.LoadX509KeyPair("/certs/cert", "/certs/key")
	if err != nil {
		panic(err)
	}

	logger.Printf("starting webhook server, listening on port 8080")
	http.HandleFunc("/mutate", mutateObject)
	server := http.Server{
		Addr: fmt.Sprintf(":%d", 8080),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		ErrorLog: logger,
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		panic(err)
	}
}

func mutateObject(w http.ResponseWriter, r *http.Request) {
	logger.Printf("received request on /mutate")

	deserializer := codecs.UniversalDeserializer()

	// Parse the AdmissionReview from the http request.
	admissionReviewRequest, err := admissionReviewFromRequest(r, deserializer)
	if err != nil {
		msg := fmt.Sprintf("error getting admission review from request: %v", err)
		logger.Printf(msg)
		w.WriteHeader(400)
		w.Write([]byte(msg))
		return
	}

	logger.Printf("mutation request for object kind: %s", admissionReviewRequest.Request.Resource.Resource)
	logger.Printf("using admission review request username, %s, to determine appropriate labels to add to object", admissionReviewRequest.Request.UserInfo.Username)

	// rawRequest := admissionReviewRequest.Request.Object.Raw

	admissionResponse := &admissionv1.AdmissionResponse{}
	patchType := admissionv1.PatchTypeJSONPatch
	labels := map[string]string{
		"sea":  "turtle",
		"land": "mongoose",
		"air":  "falcon",
	}

	patch, _ := patchBuilder(labels)

	// todo: handle resources with a pod template spec and patch those labels as well?
	// switch admissionReviewRequest.Request.Resource.Resource {
	// case "pods":
	// 	pod := corev1.Pod{}
	// 	if _, _, err := deserializer.Decode(rawRequest, nil, &pod); err != nil {
	// 		msg := fmt.Sprintf("error decoding raw pod: %v", err)
	// 		logger.Printf(msg)
	// 		w.WriteHeader(500)
	// 		w.Write([]byte(msg))
	// 		return
	// 	}
	// 	if len(pod.Labels) == 0 {
	// 		labelsExist := false
	// 		patch = buildPatch(labelsExist, labels)
	// 	} else {
	// 		labelsExist := true
	// 		patch = buildPatch(labelsExist, labels)
	// 	}
	// case "configmaps":
	// 	configmap := corev1.ConfigMap{}
	// 	if _, _, err := deserializer.Decode(rawRequest, nil, &configmap); err != nil {
	// 		msg := fmt.Sprintf("error decoding raw configmap: %v", err)
	// 		logger.Printf(msg)
	// 		w.WriteHeader(500)
	// 		w.Write([]byte(msg))
	// 		return
	// 	}
	// 	if len(configmap.Labels) == 0 {
	// 		labelsExist := false
	// 		patch = buildPatch(labelsExist, labels)
	// 	} else {
	// 		labelsExist := true
	// 		patch = buildPatch(labelsExist, labels)
	// 	}
	// default:
	// 	logger.Printf("unhandled resource type: %s\n", admissionReviewRequest.Request.Resource.Resource)
	// }

	// admissionResponse.Allowed = true
	// if patch != "" {
	// 	admissionResponse.PatchType = &patchType
	// 	admissionResponse.Patch = []byte(patch)
	// }

	// patchBytes, err := json.Marshal(patch)
	// if err != nil {
	// 	http.Error(w, "could not marshal patch bytes", http.StatusInternalServerError)
	// 	return
	// }

	// logger.Printf("patching object with following patch string: %s", string(patchBytes))

	admissionResponse.Allowed = true
	admissionResponse.PatchType = &patchType
	admissionResponse.Patch = patch

	// Construct the response, which is just another AdmissionReview.
	var admissionReviewResponse admissionv1.AdmissionReview
	admissionReviewResponse.Response = admissionResponse
	admissionReviewResponse.SetGroupVersionKind(admissionReviewRequest.GroupVersionKind())
	admissionReviewResponse.Response.UID = admissionReviewRequest.Request.UID

	resp, err := json.Marshal(admissionReviewResponse)
	if err != nil {
		msg := fmt.Sprintf("error marshalling response json: %v", err)
		logger.Printf(msg)
		w.WriteHeader(500)
		w.Write([]byte(msg))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func admissionReviewFromRequest(r *http.Request, deserializer runtime.Decoder) (*admissionv1.AdmissionReview, error) {
	// Validate that the incoming content type is correct.
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("expected application/json content-type")
	}

	// Get the body data, which will be the AdmissionReview
	// content for the request.
	var body []byte
	if r.Body != nil {
		requestData, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		body = requestData
	}

	// Decode the request body into
	admissionReviewRequest := &admissionv1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, admissionReviewRequest); err != nil {
		return nil, err
	}

	return admissionReviewRequest, nil
}

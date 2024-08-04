package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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
	logger.Printf("admission review request username: %s", admissionReviewRequest.Request.UserInfo.Username)

	// If the admission review request is coming from a system user (e.g. replicaset controller), we don't want to mutate the object in the request
	if strings.Contains(admissionReviewRequest.Request.UserInfo.Username, "system") {
		logger.Printf("ignoring admission review request because it came from a system user: %s", admissionReviewRequest.Request.UserInfo.Username)
		admissionReviewResponse := passingAdmissionReviewResponse(admissionReviewRequest)
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
	} else {
		// The admission review request must be coming from a non-system user, we do want to mutate the object in the request
		rawRequest := admissionReviewRequest.Request.Object.Raw
		k8sObject := GenericK8sObject{}
		if err := json.Unmarshal(rawRequest, &k8sObject); err != nil {
			msg := fmt.Sprintf("error unmarshalling raw request object: %v", err)
			logger.Printf(msg)
			w.WriteHeader(500)
			w.Write([]byte(msg))
			return
		}

		patch, _ := patchBuilder(k8sObject)

		admissionReviewResponse := patchingAdmissionReviewResponse(admissionReviewRequest, patch)
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

build:
	docker build -t webhook:v0 .
	kind load docker-image webhook:v0 --name space-debris

deploy-webhook:
	kubectl delete -f deploy/deployment.yaml
	kubectl apply -f deploy/deployment.yaml
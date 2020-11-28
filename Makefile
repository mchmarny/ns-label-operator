APP_VERSION  ?=v0.4.12
APP_ID       ?=ns-label-operator
IMAGE_OWNER  ?=$(shell git config --get user.username)

.PHONY: all
all: help

.PHONY: setup
setup: ## Sets up k8s cluster for tests 
	kubectl cluster-info
	kubectl get ns
	kubectl create ns ns-label-operator-test

.PHONY: tidy
tidy: ## Updates the go modules and vendors all dependencies 
	go mod tidy
	go mod vendor

.PHONY: test
test: tidy ## Tests the entire project 
	go test -v -count=1 -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: testcover
testcover: ## Tests coverage configuration 
	curl --data-binary @.codecov.yml https://codecov.io/validate

.PHONY: build
build: tidy ## build code
	go build -mod vendor -o ./bin/$(APP_ID) ./cmd/.

.PHONY: run
run: ## Runs compiled code
	KUBECONFIG=~/.kube/config \
	DEBUG=false \
	LOG_TO_JSON=true \
	CONFIG_DIR=manifests \
	TRIGGER_LABEL=dapr-enabled \
	./bin/$(APP_ID)

.PHONY: image
image: tidy ## Builds and publish image 
	docker build \
		-t ghcr.io/$(IMAGE_OWNER)/$(APP_ID):$(APP_VERSION) .
	docker push ghcr.io/$(IMAGE_OWNER)/$(APP_ID):$(APP_VERSION)

.PHONY: lint
lint: ## Lints the entire project 
	golangci-lint run --timeout=3m

.PHONY: deploy
deploy: ## Deploys pre-build image to k8s
	kubectl apply -f deployment.yaml

.PHONY: spell 
spell: ## Checks spelling across the entire project 
	# go get github.com/client9/misspell/cmd/misspell
	misspell -locale="US" -error -source="text" cmd/*
	misspell -locale="US" -error -source="text" pkg/**
	misspell -locale="US" -error -source="text" *.md

.PHONY: tag
tag: ## Creates release tag 
	git tag $(APP_VERSION)
	git push origin $(APP_VERSION)

.PHONY: helm
helm: ## Install Helm chart
	helm install trace-exporter chart/ \
		-n dev \
		--set debug=true \
		--set logAsJson=false \
		--set triggerLabel=trace-export-enabled \
		--set manifestConfigMap=trace-exporter-config

.PHONY: helm-clean
helm-clean: ## Uninstall Helm chart
	helm uninstall trace-exporter -n dev

.PHONY: clean
clean: ## Cleans up generated files 
	go clean
	rm -fr ./bin
	rm -fr ./vendor

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

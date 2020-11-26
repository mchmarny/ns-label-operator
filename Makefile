APP_VERSION  ?=v0.2.1
APP_ID       ?=ns-label-operator
IMAGE_OWNER  ?=$(shell git config --get user.username)

.PHONY: all
all: help

.PHONY: tidy
tidy: ## Updates the go modules and vendors all dependencies 
	go mod tidy
	go mod vendor

.PHONY: test
test: tidy ## Tests the entire project 
	go test -v -count=1 -race ./...

.PHONY: run
run: tidy ## Runs uncompiled code
	KUBECONFIG=~/.kube/config DEBUG=true YAML_PATH=test.yaml \
		go run cmd.go handler.go main.go

.PHONY: image
image: tidy ## Builds and publish image 
	docker build -t ghcr.io/$(IMAGE_OWNER)/$(APP_ID):$(APP_VERSION) .
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
	misspell -locale="US" -error -source="text" *.go

tag: ## Creates release tag 
	git tag $(APP_VERSION)
	git push origin $(APP_VERSION)

.PHONY: clean
clean: ## Cleans up generated files 
	go clean
	rm -fr ./bin
	rm -fr ./vendor

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

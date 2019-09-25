REGISTRY = custompodautoscaler
NAME = custom-pod-autoscaler
VERSION = latest

default: vendor
	@echo "=============Building============="
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -o dist/$(NAME)

lint: vendor
	@echo "=============Linting============="
	go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

test: vendor
	@echo "=============Running unit tests============="
	go test ./...

vendor:
	go mod vendor

docker: default
	@echo "=============Building docker images============="
	docker build --target=python -t $(REGISTRY)/python:$(VERSION) .
	docker build --target=alpine -t $(REGISTRY)/alpine:$(VERSION) .
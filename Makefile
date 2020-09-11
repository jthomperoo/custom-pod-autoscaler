REGISTRY = custompodautoscaler
NAME = custom-pod-autoscaler
VERSION = latest

default: vendor
	@echo "=============Building============="
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -o dist/$(NAME) ./cmd/custom-pod-autoscaler
	cp LICENSE dist/LICENSE

unittest: vendor
	@echo "=============Running unit tests============="
	CGO_ENABLED=0 GOOS=linux go test -mod vendor ./... -cover -coverprofile unit_cover.out --tags=unit

lint: vendor
	@echo "=============Linting============="
	go list -mod vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

docker: default
	@echo "=============Building docker images============="
	docker build --target=python-3-6 -t $(REGISTRY)/python-3-6:$(VERSION) .
	docker build --target=python-3-7 -t $(REGISTRY)/python-3-7:$(VERSION) .
	docker build --target=python-3-8 -t $(REGISTRY)/python-3-8:$(VERSION) .
	docker build --target=alpine -t $(REGISTRY)/alpine:$(VERSION) .
	docker tag $(REGISTRY)/python-3-8:$(VERSION) $(REGISTRY)/python:$(VERSION)

doc:
	@echo "=============Serving docs============="
	mkdocs serve

vendor:
	go mod vendor

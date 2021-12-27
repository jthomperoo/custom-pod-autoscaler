REGISTRY = custompodautoscaler
NAME = custom-pod-autoscaler
VERSION = latest

default: vendor_modules
	@echo "=============Building============="
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.Version=$(VERSION)'" -mod vendor -o dist/$(NAME) main.go
	cp LICENSE dist/LICENSE

test: vendor_modules
	@echo "=============Running unit tests============="
	go test -mod vendor ./... -cover -coverprofile unit_cover.out

lint: vendor_modules
	@echo "=============Linting============="
	go list -mod vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

beautify: vendor_modules
	@echo "=============Beautifying============="
	gofmt -s -w .
	go mod tidy

docker: default
	@echo "=============Building docker images============="
	docker build --target=python-3-6 -t $(REGISTRY)/python-3-6:$(VERSION) .
	docker build --target=python-3-7 -t $(REGISTRY)/python-3-7:$(VERSION) .
	docker build --target=python-3-8 -t $(REGISTRY)/python-3-8:$(VERSION) .
	docker build --target=alpine -t $(REGISTRY)/alpine:$(VERSION) .
	docker build --target=openjdk-11 -t $(REGISTRY)/openjdk-11:$(VERSION) .
	docker tag $(REGISTRY)/python-3-8:$(VERSION) $(REGISTRY)/python:$(VERSION)

doc:
	@echo "=============Serving docs============="
	mkdocs serve

view_coverage:
	@echo "=============Loading coverage HTML============="
	go tool cover -html=unit_cover.out

vendor_modules:
	go mod vendor

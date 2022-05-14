REGISTRY = custompodautoscaler
NAME = custom-pod-autoscaler
VERSION = latest

default:
	@echo "=============Building============="
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/$(NAME) main.go
	cp LICENSE dist/LICENSE

test:
	@echo "=============Running tests============="
	go test ./... -cover -coverprofile coverage.out

lint:
	@echo "=============Linting============="
	staticcheck ./...

format:
	@echo "=============Formatting============="
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

coverage:
	@echo "=============Loading coverage HTML============="
	go tool cover -html=coverage.out

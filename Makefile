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
	go run honnef.co/go/tools/cmd/staticcheck@v0.4.6 ./...

format:
	@echo "=============Formatting============="
	gofmt -s -w .
	go mod tidy

docker: default
	@echo "=============Building docker images============="
	docker ps
	docker build --target=python-3-8 -t $(REGISTRY)/python-3-8:$(VERSION) .
	docker build --target=python-3-12 -t $(REGISTRY)/python-3-12:$(VERSION) .
	docker build --target=alpine -t $(REGISTRY)/alpine:$(VERSION) .
	docker tag $(REGISTRY)/python-3-12:$(VERSION) $(REGISTRY)/python:$(VERSION)

doc:
	@echo "=============Serving docs============="
	mkdocs serve

coverage:
	@echo "=============Loading coverage HTML============="
	go tool cover -html=coverage.out

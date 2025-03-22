REGISTRY = custompodautoscaler
NAME = custom-pod-autoscaler
VERSION = latest

default: package_linux_amd64 package_linux_arm64
	docker build --target=python-3-13 --tag $(REGISTRY)/python-3-13:$(VERSION) --tag $(REGISTRY)/python:$(VERSION) .
	docker build --target=python-3-12 --tag $(REGISTRY)/python-3-12:$(VERSION) .
	docker build --target=alpine --tag $(REGISTRY)/alpine:$(VERSION) .

test:
	@echo "=============Running tests============="
	go test ./... -cover -coverprofile coverage.out

lint:
	@echo "=============Linting============="
	go run honnef.co/go/tools/cmd/staticcheck@v0.6.0 ./...

format:
	@echo "=============Formatting============="
	gofmt -s -w .
	go mod tidy

doc:
	@echo "=============Serving docs============="
	mkdocs serve

coverage:
	@echo "=============Loading coverage HTML============="
	go tool cover -html=coverage.out

package_all: package_linux_386 package_linux_amd64 package_linux_arm package_linux_arm64 package_darwin_amd64 package_darwin_arm64 package_windows_386 package_windows_amd64
	cp custom-pod-autoscaler-linux-amd64.tar.gz custom-pod-autoscaler.tar.gz

package_linux_386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/linux_386/custom-pod-autoscaler main.go
	cp LICENSE dist/linux_386/LICENSE
	tar -czvf custom-pod-autoscaler-linux-386.tar.gz dist/linux_386/*

package_linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/linux_amd64/custom-pod-autoscaler main.go
	cp LICENSE dist/linux_amd64/LICENSE
	tar -czvf custom-pod-autoscaler-linux-amd64.tar.gz dist/linux_amd64/*

package_linux_arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/linux_arm/custom-pod-autoscaler main.go
	cp LICENSE dist/linux_arm/LICENSE
	tar -czvf custom-pod-autoscaler-linux-arm.tar.gz dist/linux_arm/*

package_linux_arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/linux_arm64/custom-pod-autoscaler main.go
	cp LICENSE dist/linux_arm64/LICENSE
	tar -czvf custom-pod-autoscaler-linux-arm64.tar.gz dist/linux_arm64/*

package_darwin_amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/darwin_amd64/custom-pod-autoscaler main.go
	cp LICENSE dist/darwin_amd64/LICENSE
	tar -czvf custom-pod-autoscaler-darwin-amd64.tar.gz dist/darwin_amd64/*

package_darwin_arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/darwin_arm64/custom-pod-autoscaler main.go
	cp LICENSE dist/darwin_arm64/LICENSE
	tar -czvf custom-pod-autoscaler-darwin-arm64.tar.gz dist/darwin_arm64/*

package_windows_386:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/windows_386/custom-pod-autoscaler.exe main.go
	cp LICENSE dist/windows_386/LICENSE
	tar -czvf custom-pod-autoscaler-windows-386.tar.gz dist/windows_386/*

package_windows_amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o dist/windows_amd64/custom-pod-autoscaler.exe main.go
	cp LICENSE dist/windows_amd64/LICENSE
	tar -czvf custom-pod-autoscaler-windows-amd64.tar.gz dist/windows_amd64/*

docker_multi_platform: package_linux_amd64 package_linux_arm64
	docker buildx build --push --platform=linux/amd64,linux/arm64 --target=python-3-13 --tag $(REGISTRY)/python-3-13:$(VERSION) --tag $(REGISTRY)/python:$(VERSION) .
	docker buildx build --push --platform=linux/amd64,linux/arm64 --target=python-3-12 --tag $(REGISTRY)/python-3-12:$(VERSION) .
	docker buildx build --push --platform=linux/amd64,linux/arm64 --target=alpine --tag $(REGISTRY)/alpine:$(VERSION) .

docker_tag_latest:
	docker buildx imagetools create --tag $(REGISTRY)/python:$(VERSION) $(REGISTRY)/python:latest
	docker buildx imagetools create --tag $(REGISTRY)/python-3-13:$(VERSION) $(REGISTRY)/python-3-13:latest
	docker buildx imagetools create --tag $(REGISTRY)/python-3-13:$(VERSION) $(REGISTRY)/python-3-12:latest
	docker buildx imagetools create --tag $(REGISTRY)/alpine:$(VERSION) $(REGISTRY)/alpine:latest

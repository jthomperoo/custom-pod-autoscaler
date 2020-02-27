IMAGE = simple-pod-metrics-golang:latest

default:
	@echo "=============Building============="
	CGO_ENABLED=0 GOOS=linux go build -o dist/simple-pod-metrics-golang

docker: default
	@echo "=============Building docker images============="
	docker build -t $(IMAGE) .
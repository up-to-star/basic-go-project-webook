.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=amd64 go build -o webook .
	@docker rmi -f up2star/webook:v0.0.1
	@docker build -t up2star/webook:v0.0.1 .

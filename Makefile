VERSION := 0.3

build:
	docker buildx build \
	--platform linux/amd64 \
	-t cnk3x/ssdpd:latest \
	-t cnk3x/ssdpd:$(VERSION) \
	--build-arg  HTTP_PROXY=http://192.168.31.10:7890 \
	--build-arg HTTPS_PROXY=http://192.168.31.10:7890 \
	--load \
	.

push:
	docker buildx build \
	--platform linux/amd64,linux/arm/v7,linux/arm64/v8 \
	-t cnk3x/ssdpd:latest \
	-t cnk3x/ssdpd:$(VERSION) \
	--push \
	.

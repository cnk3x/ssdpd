build:
	docker buildx build \
	--platform linux/amd64,linux/i386,linux/arm/v7,linux/arm64/v8 \
	-t doubledong/hello \
	-t cnk3x/ssdpd:latest \
	-t cnk3x/ssdpd:0.3 . \
	--push

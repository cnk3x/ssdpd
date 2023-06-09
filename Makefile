VERSION := 0.3
ALIREPO := registry.cn-shenzhen.aliyuncs.com/

build:
	docker buildx build --platform linux/amd64 -t cnk3x/ssdpd:latest --load .

push:
	docker buildx build \
	--platform linux/amd64,linux/arm/v7,linux/arm64/v8 \
	-t cnk3x/ssdpd:latest \
	-t cnk3x/ssdpd:$(VERSION) \
	-t $(ALIREPO)cnk3x/ssdpd:latest \
	-t $(ALIREPO)cnk3x/ssdpd:$(VERSION) \
	--push \
	.

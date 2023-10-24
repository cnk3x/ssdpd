VERSION := 0.3.2
ALIREPO := registry.cn-shenzhen.aliyuncs.com/

build:
	docker build --platform linux/amd64 -t cnk3x/ssdpd:latest --load .

push:
	docker build \
	--platform linux/amd64,linux/arm/v7,linux/arm64/v8 \
	-t cnk3x/ssdpd:latest \
	-t cnk3x/ssdpd:$(VERSION) \
	-t $(ALIREPO)cnk3x/ssdpd:latest \
	-t $(ALIREPO)cnk3x/ssdpd:$(VERSION) \
	--push \
	--build-arg http_proxy=http://192.168.31.31:7890 \
	--build-arg https_proxy=http://192.168.31.31:7890 \
	.

binBuild:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -trimpath -ldflags '-s -w' -o ./bin/ssdpd-linux-amd64 ./cmd/ssdpd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -v -a -trimpath -ldflags '-s -w' -o ./bin/ssdpd-linux-arm64 ./cmd/ssdpd

FROM golang:alpine as build

WORKDIR /build

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn

COPY ./cmd/ssdpd/go.mod ./cmd/ssdpd/go.sum ./cmd/ssdpd/
RUN cd ./cmd/ssdpd/ && go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -trimpath -ldflags '-s -w' -o /ssdpd ./cmd/ssdpd

FROM busybox:stable

COPY --from=build /ssdpd /ssdpd

CMD [ "/sspdp" ]

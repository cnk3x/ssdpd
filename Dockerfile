FROM golang:alpine as build

WORKDIR /build

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn

ADD . .
RUN go build -trimpath -ldflags '-s -w' -o /ssdpd ./cmd/ssdpd

FROM busybox:stable

COPY --from=build /ssdpd /ssdpd

CMD [ "/sspdp" ]

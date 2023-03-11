FROM golang:alpine as build
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk --no-cache add upx

ENV GO111MODULE=on GOPROXY=https://goproxy.cn

WORKDIR /build/cmd/ssdpd
COPY ./cmd/ssdpd/go.mod ./cmd/ssdpd/go.sum ./
RUN go mod download

WORKDIR /build
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .

WORKDIR /build/cmd/ssdpd
RUN CGO_ENABLED=0 go build -v -trimpath -ldflags '-s -w' -o /ssdpd ./

RUN upx /ssdpd

FROM alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
COPY --from=build /ssdpd /ssdpd
CMD [ "/sspdp" ]

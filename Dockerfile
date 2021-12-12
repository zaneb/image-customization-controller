FROM quay.io/projectquay/golang:1.16 AS builder
WORKDIR /go/src/github.com/openshift/image-customization-controller
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -a -o bin/image-customization-controller cmd/controller/main.go
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -a -o bin/image-customization-server cmd/static-server/main.go

FROM quay.io/centos/centos:stream8
COPY --from=builder /go/src/github.com/openshift/image-customization-controller/bin/image-customization-controller /
COPY --from=builder /go/src/github.com/openshift/image-customization-controller/bin/image-customization-server /
RUN dnf install -y nmstate

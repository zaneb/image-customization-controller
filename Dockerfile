FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.16-openshift-4.10 AS builder
WORKDIR /go/src/github.com/openshift/image-customization-controller
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -a -o bin/image-customization-controller cmd/controller/main.go
RUN CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -a -o bin/image-customization-server cmd/static-server/main.go

FROM registry.ci.openshift.org/ocp/4.10:base
COPY --from=builder /go/src/github.com/openshift/image-customization-controller/bin/image-customization-controller /
COPY --from=builder /go/src/github.com/openshift/image-customization-controller/bin/image-customization-server /
RUN dnf install -y nmstate

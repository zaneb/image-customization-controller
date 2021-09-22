FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.16-openshift-4.9 AS builder
WORKDIR /go/src/github.com/openshift/image-customization-controller
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=mod -a -o bin/image-customization-controller main.go

FROM registry.ci.openshift.org/ocp/4.9:base
COPY --from=builder /go/src/github.com/openshift/image-customization-controller/bin/image-customization-controller /
RUN yum install -y nmstate

################################
# STEP 1 build executable binary
################################

FROM golang:1.15.6-alpine AS builder
ENV GOPATH "/go:/go/src/pv-provisioner"

RUN apk --update add git openssh gcc make g++ pkgconfig zlib-dev bash

RUN go get -u k8s.io/client-go/...
RUN go get -u golang.org/x/oauth2/google
RUN go get -u google.golang.org/api/compute/v1
RUN go get -u github.com/aws/aws-sdk-go/...

ADD src /go/src/pv-provisioner/src
WORKDIR /go/src/pv-provisioner/src

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o /usr/sbin/pv-provisioner /go/src/pv-provisioner/src/cmd/pv-provisioner.go
############################
# STEP 2 build a small image
############################
FROM scratch

COPY --from=builder /usr/sbin/pv-provisioner /usr/sbin/pv-provisioner

ENTRYPOINT ["/usr/sbin/pv-provisioner"]
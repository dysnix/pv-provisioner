FROM golang:1.15.6-alpine
ENV GOPATH "/go:/go/src/pv-provisioner"

RUN apk --update add git openssh gcc make g++ pkgconfig zlib-dev bash

RUN go get -u k8s.io/client-go/...
RUN go get -u golang.org/x/oauth2/google
RUN go get -u google.golang.org/api/compute/v1
RUN go get -u github.com/aws/aws-sdk-go/...

ADD src /go/src/pv-provisioner/src
WORKDIR /go/src/pv-provisioner/src

RUN go build -o /usr/sbin/pv-provisioner /go/src/pv-provisioner/src/cmd/pv-provisioner.go

RUN rm -rf /go/src
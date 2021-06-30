FROM golang:1.16-alpine as builder

RUN apk --update add git openssh gcc make g++ pkgconfig zlib-dev bash ca-certificates
RUN update-ca-certificates

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o /usr/sbin/pv-provisioner ./cmd/pv-provisioner/main.go

############################
# STEP 2 build a small image
############################
FROM scratch

# copy the ca-certificate.crt from the build stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/sbin/pv-provisioner /usr/sbin/pv-provisioner

ENTRYPOINT ["/usr/sbin/pv-provisioner"]
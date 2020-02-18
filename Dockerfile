############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache ca-certificates git tzdata
WORKDIR $GOPATH/src/github.com/gdgtoledo/cansino

COPY . .

# Build the binary.
RUN GOOS=linux GOARCH=386 go build -ldflags="-w -s" -o /go/bin/cansino
############################
# STEP 2 build a small image
############################
FROM scratch
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Europe/Madrid
# Copy default certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=builder /go/bin/cansino /go/bin/cansino
# Run the cansino binary.
ENTRYPOINT ["/go/bin/cansino"]
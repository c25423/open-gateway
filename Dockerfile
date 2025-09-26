# Build stage

FROM --platform=$BUILDPLATFORM golang:alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/open-gateway
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /go/bin/open-gateway ./cmd/server

# Final stage

FROM alpine:latest

RUN mkdir -p /etc/open-gateway /var/lib/open-gateway
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/open-gateway /usr/local/bin/open-gateway

CMD ["/usr/local/bin/open-gateway"]

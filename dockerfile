# Stage 1: Build
FROM golang:1.25.6-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build \
	-ldflags="-w -s" \
	-o iris-server ./cmd/iris-server



# Stage 2: Runtime
FROM alpine:latest AS runner
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

COPY --from=builder /build/iris-server .

RUN adduser -D -u 1000 iris
RUN chown -R iris:iris /build
USER iris

EXPOSE 8080
CMD ["./iris-server"]

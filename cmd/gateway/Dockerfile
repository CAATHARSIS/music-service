FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /app

RUN printf "go 1.22\n\nuse (\n\t./api/gen\n\t./cmd/gateway\n\t./internal/gateway\n\t./pkg\n)\n" > go.work

COPY api/gen/go.mod api/gen/go.mod
COPY api/gen/go.sum api/gen/go.sum
COPY cmd/gateway/go.mod cmd/gateway/go.mod
# COPY cmd/gateway/go.sum cmd/gateway/go.sum
COPY internal/gateway/go.mod internal/gateway/go.mod
COPY internal/gateway/go.sum internal/gateway/go.sum
COPY pkg/go.mod pkg/go.mod
COPY pkg/go.sum pkg/go.sum

RUN go work sync && cd cmd/gateway && go mod download

COPY api/gen api/gen
COPY internal/gateway/ internal/gateway/
COPY cmd/gateway/ cmd/gateway/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/gateway ./cmd/gateway

FROM alpine AS runner
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /app/gateway .
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
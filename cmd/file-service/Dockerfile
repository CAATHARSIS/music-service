FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

RUN printf "go 1.26\n\nuse (\n\t./api/gen\n\t./cmd/file-service\n\t./internal/file\n\t./pkg\n)" > go.work

COPY api/gen/go.mod api/gen/go.mod
COPY api/gen/go.sum api/gen/go.sum
COPY cmd/file-service/go.mod cmd/file-service/go.mod
COPY cmd/file-service/go.sum cmd/file-service/go.sum
COPY internal/file/go.mod internal/file/go.mod
COPY internal/file/go.sum internal/file/go.sum
COPY pkg/go.mod pkg/go.mod
COPY pkg/go.sum pkg/go.sum

RUN go work sync && cd cmd/file-service && go mod download

COPY api/gen api/gen
COPY internal/file/ internal/file/
COPY cmd/file-service/ cmd/file-service/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/file-service ./cmd/file-service

FROM alpine AS runner

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/file-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 50053

ENTRYPOINT [ "/app/file-service" ]
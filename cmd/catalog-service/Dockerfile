FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

RUN printf "go 1.26\n\nuse (\n\t./api/gen\n\t./cmd/catalog-service\n\t./internal/catalog\n\t./pkg\n)" > go.work

COPY api/gen/go.mod api/gen/go.mod
COPY api/gen/go.sum api/gen/go.sum
COPY cmd/catalog-service/go.mod cmd/catalog-service/go.mod
COPY cmd/catalog-service/go.sum cmd/catalog-service/go.sum
COPY internal/catalog/go.mod internal/catalog/go.mod
COPY internal/catalog/go.sum internal/catalog/go.sum
COPY pkg/go.mod pkg/go.mod
COPY pkg/go.sum pkg/go.sum

RUN go work sync && cd cmd/catalog-service && go mod download

COPY api/gen api/gen
COPY internal/catalog/ internal/catalog/
COPY cmd/catalog-service/ cmd/catalog-service/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/catalog-service ./cmd/catalog-service

FROM alpine AS runner

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/catalog-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 50053

ENTRYPOINT [ "/app/catalog-service" ]
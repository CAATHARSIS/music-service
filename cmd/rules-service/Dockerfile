FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.work ./
RUN printf "go 1.26\n\nuse (\n\t./api/gen\n\t./cmd/rules-service\n\t./internal/rules\n\t./pkg\n)" > go.work

COPY api/gen/go.mod api/gen/go.mod
COPY api/gen/go.sum api/gen/go.sum
COPY cmd/rules-service/go.mod cmd/rules-service/go.mod
COPY cmd/rules-service/go.sum cmd/rules-service/go.sum
COPY internal/rules/go.mod internal/rules/go.mod
COPY internal/rules/go.sum internal/rules/go.sum
COPY pkg/go.mod pkg/go.mod
COPY pkg/go.sum pkg/go.sum

RUN go work sync && cd cmd/rules-service && go mod download

COPY api/gen api/gen
COPY internal/rules/ internal/rules/
COPY cmd/rules-service/ cmd/rules-service/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/rules-service ./cmd/rules-service

FROM alpine AS runner

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/rules-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 50055

ENTRYPOINT [ "/app/rules-service" ]
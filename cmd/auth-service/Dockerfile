FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.work ./
RUN printf "go 1.26\n\nuse (\n\t./api/gen\n\t./cmd/auth-service\n\t./internal/auth\n\t./pkg\n)" > go.work

COPY api/gen/go.mod api/gen/go.mod
COPY api/gen/go.sum api/gen/go.sum
COPY cmd/auth-service/go.mod cmd/auth-service/go.mod
COPY cmd/auth-service/go.sum cmd/auth-service/go.sum
COPY internal/auth/go.mod internal/auth/go.mod
COPY internal/auth/go.sum internal/auth/go.sum
COPY pkg/go.mod pkg/go.mod
COPY pkg/go.sum pkg/go.sum

RUN go work sync && cd cmd/auth-service && go mod download

COPY api/gen api/gen
COPY internal/auth/ internal/auth/
COPY cmd/auth-service/ cmd/auth-service/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/auth-service ./cmd/auth-service

FROM alpine AS runner

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/auth-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 50051

ENTRYPOINT [ "/app/auth-service" ]
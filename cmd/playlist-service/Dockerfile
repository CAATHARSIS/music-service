FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.work ./
RUN printf "go 1.26\n\nuse (\n\t./api/gen\n\t./cmd/playlist-service\n\t./internal/playlist\n\t./pkg\n)" > go.work

COPY api/gen/go.mod api/gen/go.mod
COPY api/gen/go.sum api/gen/go.sum
COPY cmd/playlist-service/go.mod cmd/playlist-service/go.mod
COPY cmd/playlist-service/go.sum cmd/playlist-service/go.sum
COPY internal/playlist/go.mod internal/playlist/go.mod
COPY internal/playlist/go.sum internal/playlist/go.sum
COPY pkg/go.mod pkg/go.mod
COPY pkg/go.sum pkg/go.sum

RUN go work sync && cd cmd/playlist-service && go mod download

COPY api/gen api/gen
COPY internal/playlist/ internal/playlist/
COPY cmd/playlist-service/ cmd/playlist-service/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/playlist-service ./cmd/playlist-service

FROM alpine AS runner

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/playlist-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 50051

ENTRYPOINT [ "/app/playlist-service" ]
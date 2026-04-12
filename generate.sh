#!/usr/bin/env bash
set -e

PROTO_DIR="api/proto"
BASE_OUT_DIR="api/gen"
GOPATH_BIN="$(go env GOPATH)/bin"

PROTOC="protoc"
GO_PLUGIN="$GOPATH_BIN/protoc-gen-go.exe"
GRPC_PLUGIN="$GOPATH_BIN/protoc-gen-go-grpc.exe"

if [ $# -eq 0 ]; then
    echo ".proto файлы не указаны"
    exit 1
fi

for arg in "$@"; do
    # Нормализуем имя: отбрасываем путь, если передан ./api/proto/file.proto
    NAME=$(basename "$arg" .proto)
    FILE="$PROTO_DIR/$NAME.proto"
    
    if [ ! -f "$FILE" ]; then
        echo "Файл не найден: $arg"
        continue
    fi

    FILE_OUT_DIR="$BASE_OUT_DIR/$NAME"
    echo "Генерация: $NAME.proto -> $FILE_OUT_DIR/"
    mkdir -p "$FILE_OUT_DIR"

    $PROTOC \
      --proto_path="$PROTO_DIR" \
      --plugin=protoc-gen-go="$GO_PLUGIN" \
      --go_out="$FILE_OUT_DIR" --go_opt=paths=source_relative \
      --plugin=protoc-gen-go-grpc="$GRPC_PLUGIN" \
      --go-grpc_out="$FILE_OUT_DIR" --go-grpc_opt=paths=source_relative \
      "$FILE"

    echo "$NAME.proto готов"
done
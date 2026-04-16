FROM golang:1.25-alpine AS builder

ARG BIN_NAME="go-fx-template"
ARG MAIN_PATH="cmd/main.go"
ARG GO_BUILD_FLAGS="-s -w"
ARG UPX_FLAGS="--best --lzma"

# RUN apk add --no-cache upx

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="$GO_BUILD_FLAGS" -o $BIN_NAME $MAIN_PATH

# RUN upx $UPX_FLAGS $BIN_NAME

FROM alpine:3.22

ARG BIN_NAME="go-fx-template"

RUN addgroup -S app && adduser -S -G app -u 10001 app

WORKDIR /app

COPY --from=builder --chown=app:app /build/$BIN_NAME ./

USER app

EXPOSE 8080

CMD ["./go-fx-template"]

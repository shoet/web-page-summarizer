# ===== build stage ====
FROM golang:1.20.12-bullseye as builder

WORKDIR /app

COPY ./summarytask/go.mod ./summarytask/go.mod
COPY ./summarytask/go.sum ./summarytask/go.sum
COPY ./go.mod ./go.mod

RUN --mount=type=cache,target=/go-mod-cache \
    cd ./summarytask && go mod download

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    cd ./summarytask && \
    go build -trimpath -ldflags="-w -s" -tags timetzdata -o ./bin/main ./cmd/main.go

# ===== deploy stage ====
FROM golang:1.20.12-alpine3.18 as deploy

WORKDIR /app

RUN apk update

RUN apk add chromium
RUN apk add libc6-compat
RUN apk add gcompat
RUN ln -s /lib/libc.so.6 /usr/lib/libresolv.so.2

COPY --from=builder /app/summarytask/bin/main ./main

CMD ["/app/main"]

# ===== build stage ====
FROM golang:1.19.13-bullseye as builder

WORKDIR /app

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache \
    go mod download

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    go build -trimpath -ldflags="-w -s" -o functionw/bin/api functions/api/main.go

# ===== local development stage ====
FROM golang:1.19.13-bullseye as dev

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest
CMD ["air"]

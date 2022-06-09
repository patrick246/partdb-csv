FROM golang:1.18 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN mkdir -p bin && CGO_ENABLED=0 go build -ldflags '-w -s' -o ./bin ./cmd/...

FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source=https://github.com/patrick246/partdb-csv
LABEL org.opencontainers.image.authors=patrick246
WORKDIR /app
COPY --from=builder /app/bin/partdb-csv /partdb-csv

ENTRYPOINT ["/partdb-csv"]

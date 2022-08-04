FROM golang:1.19.0-alpine as builder

WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" github.com/ilyam8/periodic-pr-labeler/cmd/labeler

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/labeler /
ENTRYPOINT ["/labeler"]

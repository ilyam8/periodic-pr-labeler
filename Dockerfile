FROM golang:alpine as builder

RUN mkdir -p /workspace
WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" github.com/ilyam8/periodic-pr-labeler/cmd/labeler

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/labeler /
ENTRYPOINT ["/labeler"]
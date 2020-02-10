FROM golang as builder

RUN mkdir -p /workspace
WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 go build github.com/ilyam8/periodic-pr-labeler/cmd/labeler

FROM gcr.io/distroless/static
COPY --from=builder /workspace/labeler /
ENTRYPOINT ["/labeler"]
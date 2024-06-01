# Build stage
FROM golang:1.22 as build

WORKDIR /go/src/arcprog-4
COPY . .

RUN go test ./...
ENV CGO_ENABLED=0
RUN go install ./cmd/...

# ==== Final image ====
FROM alpine:latest
WORKDIR /opt/arcprog-4

COPY entry.sh /opt/arcprog-4/
RUN dos2unix /opt/arcprog-4/entry.sh && chmod +x /opt/arcprog-4/entry.sh

COPY --from=build /go/bin/* /opt/arcprog-4

RUN ls -l /opt/arcprog-4
ENTRYPOINT ["/opt/arcprog-4/entry.sh"]
CMD ["server"]

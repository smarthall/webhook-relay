FROM golang:1.22.3 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o relay .

FROM gcr.io/distroless/static-debian11

COPY --from=builder /app/relay /relay

CMD ["/relay"]

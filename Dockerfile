FROM golang:latest AS nbuilder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o denonapi

FROM scratch
COPY --from=builder /app/denonapi /denonapi
COPY ./index.html /index.html

ENTRYPOINT ["/denonapp"]


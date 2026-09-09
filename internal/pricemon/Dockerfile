FROM golang:1.27-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o pricemon ./cmd/pricemon

EXPOSE 8080

CMD ["./pricemon"]
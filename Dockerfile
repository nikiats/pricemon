FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
	&& CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
	&& CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dealmanager ./cmd/dealmanager \
	&& CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate-dealmanager ./cmd/migrate-dealmanager

FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=build /out/server /out/migrate /out/dealmanager /out/migrate-dealmanager ./
COPY --from=build /src/public ./public

USER app
EXPOSE 8082 8083

CMD ["/app/server"]

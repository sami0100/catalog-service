FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /bin/catalog

FROM alpine:3.20
WORKDIR /app
COPY --from=build /bin/catalog /app/catalog
EXPOSE 3002
CMD ["/app/catalog"]

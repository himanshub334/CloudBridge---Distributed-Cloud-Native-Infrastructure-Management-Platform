FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/cloudbridge .
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/cloudbridge /usr/local/bin/cloudbridge
EXPOSE 8080
CMD ["cloudbridge"]

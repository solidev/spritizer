# Stage 1 (to create a "build" image, ~850MB)
FROM golang:1.14 AS builder

RUN mkdir /spritizer
COPY . /spritizer/
WORKDIR /spritizer/src/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildmode exe -o ../spritizer .
RUN chmod +x /spritizer/spritizer

# Stage 2 (create executable image)
FROM alpine
RUN apk --no-cache add inkscape
WORKDIR /usr/bin/
COPY --from=builder /spritizer/spritizer .

ENTRYPOINT ["/usr/bin/spritizer"]
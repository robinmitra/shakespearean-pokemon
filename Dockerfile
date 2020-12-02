FROM golang:alpine AS build
WORKDIR /go/src/app
COPY . .
RUN go build -o bin/app
CMD ["bin/app"]

FROM golang:alpine AS builder
RUN apk add --no-cache --update \
        git \
        ca-certificates
ADD . /app
WORKDIR /app
COPY go.mod go.sum ./
RUN  go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /main .

# final stage
FROM alpine

WORKDIR /app
ADD public /app/public
ADD public/build /app/public/build
ADD public/js /app/public/js
ADD public/libs /app/public/libs
ADD public/libs/autocomplete /app/public/libs/autocomplete
ADD public/libs/scroll /app/public/libs/scroll
COPY --from=builder /main ./
RUN chmod +x ./main
ENTRYPOINT ["./main"]

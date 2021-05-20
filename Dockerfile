FROM golang:alpine AS builder
RUN apk add --no-cache --update \
        git \
        ca-certificates
ADD . /app
ADD public /app/public
WORKDIR /app
COPY go.mod go.sum ./
RUN  go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /main .

# final stage
FROM alpine
COPY --from=builder /main ./
RUN chmod +x ./main
ENTRYPOINT ["./main"]

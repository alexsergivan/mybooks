FROM golang:alpine AS builder
RUN apk add --no-cache --update \
        git \
        ca-certificates
ADD . /app
ADD public /app/public
ADD public/build /app/public/build
ADD public/js /app/public/js
ADD public/libs /app/public/libs
ADD public/libs/autocomplete /app/public/libs/autocomplete
ADD public/libs/scroll /app/public/libs/scroll
WORKDIR /app

COPY go.mod go.sum ./
RUN  go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /app/main .

# final stage
FROM alpine


COPY --from=builder /app/public /app/public
COPY --from=builder /app/public/build /app/public/build
COPY --from=builder /app/public/js /app/public/js
COPY --from=builder /app/public/libs /app/public/libs
COPY --from=builder /app/public/libs/autocomplete /app/public/libs/autocomplete
COPY --from=builder /app/public/libs/scroll /app/public/libs/scroll

COPY --from=builder /app/main /app/main

WORKDIR /app

RUN chmod +x main
ENTRYPOINT ["./app/main"]

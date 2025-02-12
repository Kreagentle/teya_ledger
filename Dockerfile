# legder/Dockerfile

FROM golang:1.23-alpine3.19 AS build

WORKDIR /teya_ledger

# copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# run tests
RUN go test -v ./...

# build the Ledger API service
RUN go build -o ledger_app ./cmd/ledger-server/main.go

# minimal image
FROM alpine:3.13

WORKDIR /teya_ledger

# copy the built binary from the build stage
COPY --from=build /teya_ledger/ledger_app .

EXPOSE 8000

CMD ["./ledger_app"]

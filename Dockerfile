FROM golang:alpine

COPY . /dest

WORKDIR /dest

EXPOSE 8080

RUN go build -o main ./cmd/app/main.go

CMD ["./main"]
FROM golang:1.14

WORKDIR /go/src/app

COPY . .

RUN go install -v ./...

EXPOSE 80

CMD ["edustream-server"]

FROM golang:1.14

WORKDIR /go/src/app

COPY . .

RUN go install -v ./...

EXPOSE 443

CMD ["edustream-server"]

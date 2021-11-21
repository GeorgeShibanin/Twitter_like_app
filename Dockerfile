FROM golang:latest


RUN mkdir -p /go/src/app
WORKDIR /go/src/app/
COPY . /go/src/app/

RUN go mod tidy
RUN go build /go/src/app/main.go

ENV SERVER_PORT 8080
EXPOSE $SERVER_PORT

CMD ["/go/src/app/main"]

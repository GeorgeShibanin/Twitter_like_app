FROM golang:latest


RUN mkdir -p /go/src/app
WORKDIR /go/src/app/
COPY . /go/src/app/

RUN go mod tidy
RUN go build -o app

ENV SERVER_PORT 8080
ENV MONGO_URL 'mongodb://database:27017'
ENV MONGO_DBNAME 'twitterPosts'
EXPOSE $SERVER_PORT
EXPOSE 27017

CMD ["/go/src/app/app"]

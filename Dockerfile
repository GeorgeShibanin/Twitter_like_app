FROM golang:latest


WORKDIR /go/src/app/
COPY . .

RUN go mod tidy
RUN go build -o app

ENV SERVER_PORT 8080
ENV MONGO_URL 'mongodb://localhost:27017/twitterPosts'
ENV MONGO_DBNAME 'twitterPosts'
ENV STORAGE_MODE 'memory'
EXPOSE $SERVER_PORT
EXPOSE 27017

CMD ["/go/src/app/app"]

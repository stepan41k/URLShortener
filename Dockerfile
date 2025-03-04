FROM golang:latest

RUN go version
ENV GOPATH=/

COPY ./ ./

#install psql
RUN apt-get update
RUN apt-get -y install postgresql-client

#male wait-for-postgres.sh executable
RUN chmod +x wait-for-postgres.sh
 
#build go app
RUN go mod download
RUN go build -o url-shortener-app ./cmd/url-shortener/main.go
CMD ["./url-shortener-app"]
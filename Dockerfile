FROM golang:alpine

WORKDIR /app

COPY . /app

RUN go build -o menufy cmd/school-cafe-scraper/main.go

ENTRYPOINT [ "menufy", "-c", "config.yml" ]

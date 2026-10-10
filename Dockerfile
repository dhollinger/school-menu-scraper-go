FROM golang:alpine

WORKDIR /app

COPY . /app

RUN go build -o menget cmd/school-cafe-scraper/main.go

ENTRYPOINT [ "menget", "-c", "config.yml" ]

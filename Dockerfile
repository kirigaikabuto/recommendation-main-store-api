FROM golang:1.13-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY *.env ./

RUN go build -o /work-api

EXPOSE 8000

CMD [ "/work-api", "-c", "prod.env"]
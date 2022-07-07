FROM golang:latest

WORKDIR /app

RUN mkdir -p ./frontend/dist

COPY frontend/dist/* /app/frontend/dist

COPY go.mod go.sum ./

RUN go mod download

COPY *.go .

RUN go build -o main .

EXPOSE 8000

CMD ["./main"]
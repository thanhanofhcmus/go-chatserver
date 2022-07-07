FROM golang:latest

WORKDIR /app

RUN mkdir -p ./frontend/dist/assets

COPY frontend/dist/assets/* /app/frontend/dist/assets

COPY frontend/dist/index.html /app/frontend/dist
COPY frontend/dist/favicon.ico /app/frontend/dist

COPY go.mod go.sum ./

RUN go mod download

COPY *.go .

RUN go build -o main .

EXPOSE 8000

CMD ["./main"]
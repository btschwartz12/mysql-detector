FROM golang:1.22
WORKDIR /app
COPY . .
RUN mkdir -p /app/var
RUN go build -o mysql-detector
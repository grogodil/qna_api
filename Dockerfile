FROM golang:1.23.4-alpine
WORKDIR /app
COPY . .
RUN go build -o server ./main.go
EXPOSE 8080
CMD ["./server"]
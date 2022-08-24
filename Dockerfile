FROM golang:alpine
WORKDIR /app
COPY . .
RUN go build -o main main.go

EXPOSE 8080
EXPOSE 8081

CMD [ "/app/main" ]
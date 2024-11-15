FROM golang:1.19

WORKDIR /app

COPY go.mod ./

RUN go mod download
 
COPY ./web/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /counter ./counter

EXPOSE 9993
CMD ["/counter"]

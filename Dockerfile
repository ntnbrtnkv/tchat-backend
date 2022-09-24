FROM golang:1.19

RUN apt update; \
    apt install imagemagick -y --no-install-recommends

ENV APP_HOME /app
RUN mkdir -p "$APP_HOME"
ADD ./go.mod /app/go.mod
ADD ./go.sum /app/go.sum
ADD ./main.go /app/main.go

ENV IMAGE_MAGIC=convert

WORKDIR "$APP_HOME"

RUN go build -o main .

EXPOSE 8080

CMD ["/app/main"]
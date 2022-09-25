FROM golang:1.19

ARG DEBIAN_FRONTEND=noninteractive
ARG IMAGEMAGICK_VERSION=7.1.0-47

RUN apt update \
    && apt-get --no-install-recommends -y install -y wget build-essential make pkg-config \
    && apt-get --no-install-recommends -y install libgif-dev libwebp-dev \
    && wget https://github.com/ImageMagick/ImageMagick/archive/${IMAGEMAGICK_VERSION}.tar.gz \
    && tar xvzf ${IMAGEMAGICK_VERSION}.tar.gz \
    && cd ImageMagick* \
    && ./configure \
        --without-magick-plus-plus \
        --disable-openmp \
        --with-gvc=no \
        --disable-docs \
    && make -j$(nproc) \
    && make install \
    && ldconfig /usr/local/lib \
    && rm -rf /var/lib/apt/lists/* \
    && cd .. \
    && rm -rf ImageMagick*

ENV APP_HOME /app
ENV GIN_MODE release

RUN mkdir -p "$APP_HOME"
ADD ./go.mod /app/go.mod
ADD ./go.sum /app/go.sum
ADD ./main.go /app/main.go

ENV IMAGE_MAGIC=convert

WORKDIR "$APP_HOME"

RUN go build -o main .

EXPOSE 8080

CMD ["/app/main"]
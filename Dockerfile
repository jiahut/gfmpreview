FROM golang:1.9-alpine as build-env

COPY . /go/src/github.com/vrischmann/gfmpreview
WORKDIR /go/src/github.com/vrischmann/gfmpreview

RUN go install

FROM alpine

COPY --from=build-env /go/bin/gfmpreview /bin/gfmpreview

ENTRYPOINT /bin/gfmpreview

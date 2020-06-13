FROM golang:1.14

COPY . /go/src/backend

RUN go get -d /go/src/backend
RUN go install backend

ENTRYPOINT /go/bin/backend

EXPOSE 3000
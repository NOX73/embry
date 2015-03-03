FROM golang:1.4.2

ADD . /go/src/embry
WORKDIR /go/src/embry

RUN go get

RUN go install embry

CMD /go/bin/embry

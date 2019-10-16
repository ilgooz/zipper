FROM golang:1.13.0

WORKDIR /go/src/github.com/ilgooz/structionsite
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN go test -v ./...

EXPOSE 80

CMD ["structionsite", "--addr", ":80"]
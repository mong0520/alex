FROM golang:1.14.2-alpine

ENV ALEX_PATH=$GOPATH/src/github.com/mong0520/alex
RUN apk add --no-cache make git
RUN go get -u github.com/golang/dep/cmd/dep
ADD . ${ALEX_PATH}
RUN cd ${ALEX_PATH} && go build
WORKDIR ${ALEX_PATH}
EXPOSE 8000

ENTRYPOINT ["./alex", "-c", "./config.json"]

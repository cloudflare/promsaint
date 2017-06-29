FROM golang:1.8.3-alpine

COPY . /go/src/github.com/cloudflare/promsaint

COPY ./run.sh /run.sh

RUN go install \
    github.com/cloudflare/promsaint/cmd/promsaint && \
    rm -rf /go/src

USER nobody

ENTRYPOINT ["/run.sh"]

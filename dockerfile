FROM golang:1.13-alpine3.10
MAINTAINER jedipunkz

WORKDIR /go/src/

ADD . /go/src/
ADD ~/.slack-ansible.yaml $HOME/

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/slack-ansible

ENTRYPOINT ["/go/bin/slack-ansible"]

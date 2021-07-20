FROM golang:1.16-alpine
WORKDIR /apod

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY main.go .
COPY api/ api/
COPY handlers/ handlers/
COPY nasa/ nasa/

RUN go build

CMD [ "./nasa-apod-fetcher" ]

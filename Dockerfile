FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN go build

# set these, or comment them out and set them manually
# when starting the container
ENV TEMPORAL_REGION=value
ENV TEMPORAL_NAMESPACE=value
# make sure to mount the key/cert
ENV TEMPORAL_TLS_KEY=value
ENV TEMPORAL_TLS_CERT=value

CMD [ "./temporal-mrn-worker" ]
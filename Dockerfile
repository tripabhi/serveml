FROM golang:1.20.5-alpine AS builder

WORKDIR /go/src/app
COPY go.mod go.sum ./
COPY cmd/agent/main.go .

COPY pkg ./pkg

RUN go build -o agent


FROM python:3.9

EXPOSE 9081
EXPOSE 3000

WORKDIR /user-app

COPY examples/simple-torch-inference/requirements.txt ./requirements.txt

RUN pip install --no-cache-dir --upgrade -r ./requirements.txt

COPY examples/simple-torch-inference/app ./app

COPY ./custom_init.sh ./
COPY --from=builder /go/src/app/agent /user-app/agent

CMD ["./custom_init.sh"]


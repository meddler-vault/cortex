FROM golang:1.13 as build


ARG CGO_ENABLED=0
ARG GO111MODULE="on"
ARG GOPROXY=""

WORKDIR /go/src/github.com/meddler-xyz/watchdog

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o of-watchdog . \
    && CGO_ENABLED=0 GOOS=darwin go build -a -ldflags "-s -w" -installsuffix cgo -o of-watchdog-darwin  \
    && GOARM=6 GOARCH=arm CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o of-watchdog-armhf . \
    && GOARCH=arm64 CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o of-watchdog-arm64 . \
    && GOOS=windows CGO_ENABLED=0 go build -a -ldflags "-s -w" -installsuffix cgo -o of-watchdog.exe .


CMD [ "./of-watchdog" ]
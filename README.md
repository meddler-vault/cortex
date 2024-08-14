# watchdog

<!-- Build -->
GOOS=linux GOARCH=amd64  go build  -o ./build ./


<!--  -->
<!-- Result -->
NATS_CONNECTION_STRING=nats://whitehat:4Jy6P%29%24Ep%40c%5ESenL@hawki-rabbitmq.indiatimes.com:4222
message_queue_topic=tasks_publish 
DEBUG=true publish_subject=builds_executor  uuid=2 NATS_CONNECTION_STRING=$NATS_CONNECTION_STRING   message_queue_topic=$message_queue_topic go run main.go
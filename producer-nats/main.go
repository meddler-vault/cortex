package producernats

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/meddler-vault/cortex/logger"
)

func getenvStr(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}
func failOnError(err error, msg string) {
	if err != nil {
		logger.Fatalln(msg, err)
	}
}

func Produce(username string, password string, host string, topic string, data string) (err error) {

	encodedUser := url.QueryEscape(username)
	encodedPassword := url.QueryEscape(password)

	connectionString := fmt.Sprintf("wss://%s:%s@%s", encodedUser, encodedPassword, host)

	queue := NewQueue(connectionString, topic)
	// if(err != nil){
	// 	return err
	// }
	log.Println("Queue ceated", topic)
	err = queue.Send(data)
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	queue.connection.FlushWithContext(
		ctx,
	)

	return err

}

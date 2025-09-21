// This application consumes the generated logs from the kafka topic
package main

import (
	"context"
	"fmt"
	"kafka-logging-system/internal/models"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
)

type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	//Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of the session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupSession's messages()
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	//This function is called within a goroutine
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			//Process the log message
			consumer.proccessLogMessage(message)

			//Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// Process the log message
func (consumer *Consumer) proccessLogMessage(message *sarama.ConsumerMessage) {
	//Parse the json log entry
	logEntry, err := models.FromJson(message.Value)
	if err != nil {
		fmt.Println("Error parsing the log message ", err)
		return
	}

	consumer.displayLog(logEntry, message.Partition, message.Offset)
}

func (consumer *Consumer) displayLog(entry *models.LogEntry, partition int32, offset int64) {
	//Color codes for different log levels
	colors := map[models.LogLevel]string{
		models.DEBUG: "\033[36m", // Cyan
		models.INFO:  "\033[32m", // Green
		models.WARN:  "\033[33m", // Yellow
		models.ERROR: "\033[31m", // Red
		models.FATAL: "\033[35m", // Magenta
	}

	reset := "\033[0m" // Reset color

	color, exists := colors[entry.Level]
	if !exists {
		color = reset
	}

	// Format: [TIMESTAMP] [APP] [LEVEL] MESSAGE [metadata]
	fmt.Printf("%s[%s] [%s] [%s] %s%s",
		color,
		entry.Timestamp.Format("15:04:05"),
		entry.Application,
		entry.Level,
		entry.Message,
		reset,
	)

	//Add partition and offset info
	fmt.Printf(" (p:%d, o:%d)", partition, offset)

	fmt.Println()
}

func main() {
	//Consumer group ID - multiple consumers with the same group id will share the same load
	consumerGroup := "log-consumer-group"
	topics := []string{"raw-logs"}

	//Kafka Consumer Configuration
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest //start from the beginning if no offset

	//Create consumer group client
	client, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, consumerGroup, config)
	if err != nil {
		log.Fatalln("Error creating consumerGroup client ", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	consumer := Consumer{
		ready: make(chan bool),
	}

	go func() {
		defer wg.Done()
		for {
			// "Consumer" should be called inside an infinite loop
			if err = client.Consume(context.Background(), topics, &consumer); err != nil {
				log.Panicf("Error from Consumer %v", err)
			}

			//Check if context was cancelled, signalling that the consumer should stop
			if context.Background().Err() != nil {
				return
			}

			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // await till the consumer has been Setup
	log.Printf("Consumer group %s started ,consuming topics: %v", consumerGroup, topics)
	fmt.Println("Ctrl-C to stop...")

	//Handle graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	<-sigterm
	log.Println("Terminating Consumer...")

	wg.Wait()

	if err := client.Close(); err != nil {
		log.Panicf("Error Closing client %v", err)
	}
}

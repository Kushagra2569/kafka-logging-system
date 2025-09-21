// This application produces random logs
package main

import (
	"fmt"
	"kafka-logging-system/internal/models"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/IBM/sarama"
)

type LogProducer struct {
	producer sarama.SyncProducer
	appName  string
}

func NewLogProducer(appName string) (*LogProducer, error) {
	//Kafka Configuration
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll //wait for all replicas
	config.Producer.Retry.Max = 3

	//Create producer
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer %w", err)
	}

	return &LogProducer{
		producer: producer,
		appName:  appName,
	}, nil
}

func (lp *LogProducer) generateLogEntry() *models.LogEntry {
	infoMessages := []string{
		"User logged in successfully",
		"data Retrieved successfully",
		"Service Started",
		"Request completed",
	}
	warnMessages := []string{
		"Slow Database query",
		"High Memory usage",
		"Connection Pool almost full",
		"Password failed for user",
		"Unauthenticated user trying to access data",
	}

	errorMessages := []string{
		"Database Connection Failed",
		"Invalid user credentials",
		"Backup Database not responding",
		"Service Unavailable",
		"Request Timeout",
	}

	//Randomly select log level
	levelRand := rand.Float32()
	var level models.LogLevel
	var message string

	switch {
	case levelRand < 0.6:
		level = models.INFO
		message = infoMessages[rand.Intn(len(infoMessages))]

	case levelRand < 0.8:
		level = models.WARN
		message = warnMessages[rand.Intn(len(warnMessages))]

	case levelRand < 0.95:
		level = models.ERROR
		message = errorMessages[rand.Intn(len(errorMessages))]

	default:
		level = models.DEBUG
		message = "debug trace information"
	}

	return &models.LogEntry{
		Timestamp:   time.Now(),
		Application: lp.appName,
		Level:       level,
		Message:     message,
	}
}

func (lp *LogProducer) sendLog() error {
	logentry := lp.generateLogEntry()

	//convert to json
	jsondata, err := logentry.ToJson()
	if err != nil {
		return fmt.Errorf("failed to marshal logentry %w ", err)
	}

	//create kafka message
	msg := &sarama.ProducerMessage{
		Topic:     "raw-logs",
		Key:       sarama.StringEncoder(logentry.Application),
		Value:     sarama.ByteEncoder(jsondata),
		Timestamp: logentry.Timestamp,
	}

	//send message
	partition, offset, err := lp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message %w", err)
	}

	fmt.Printf("[%s] Sent lot to partition %d, offset %d: %s - %s\n", logentry.Application, partition, offset, logentry.Level, logentry.Message)

	return nil
}

func (lp *LogProducer) Close() error {
	return lp.producer.Close()
}

func main() {
	appNames := []string{
		"userService",
		"DatabaseService",
		"AuthService",
		"PaymentService",
	}
	currentApp := appNames[rand.Intn(len(appNames))]

	//Create producer
	producer, err := NewLogProducer(currentApp)
	if err != nil {
		log.Fatal("Failed to create prdoducer %w", err)
	}

	defer producer.Close()

	//Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	//Start producing logs
	ticker := time.NewTicker(time.Second * time.Duration(rand.Intn(5)+1))
	defer ticker.Stop()

	fmt.Println("starting log prdoducer for application ", currentApp)
	fmt.Println("Press Ctrl + c to stop...")

	for {
		select {
		case <-ticker.C:
			if err := producer.sendLog(); err != nil {
				fmt.Println("Error sending log ", err)
			}
		case <-sigChan:
			fmt.Println("Shutting down Producer...")
			return
		}
	}

}

package main

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

func main() {
	brokers := []string{"kafka:9092"}
	topic := "demo-topic"

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// client, _ := sarama.NewClient(brokers, config)
	// cGroup, _ := sarama.NewConsumerGroupFromClient("group-1", client)
	// cGroup.Consume()

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume partition: %v", err)
	}
	defer partitionConsumer.Close()

	fmt.Println("Consumer started. Waiting for messages...")

	for msg := range partitionConsumer.Messages() {
		fmt.Printf("Received message: %s (partition=%d, offset=%d)\n", string(msg.Value), msg.Partition, msg.Offset)
	}
}

package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
)

func NewConsumer(hosts []string) (sarama.Consumer, error) {
	const op = "pkg.Kafka.NewConsumer"

	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(hosts, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return consumer, nil
}

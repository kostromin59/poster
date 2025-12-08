package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
)

func NewAsyncProducer(hosts []string) (sarama.AsyncProducer, error) {
	const op = "pkg.Kafka.NewAsyncProducer"

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Errors = true

	asyncProducer, err := sarama.NewAsyncProducer(hosts, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return asyncProducer, nil
}

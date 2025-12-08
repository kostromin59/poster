package dispatchers

import (
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
)

type AsyncKafka struct {
	asyncProducer sarama.AsyncProducer
	topic         string
}

func NewAsyncKakfa(asyncProducer sarama.AsyncProducer, topic string) *AsyncKafka {
	return &AsyncKafka{
		asyncProducer: asyncProducer,
		topic:         topic,
	}
}

func (ak *AsyncKafka) Dispatch(e any) {
	const op = "dispatchers.AsyncKafka.Dispatch"
	log := slog.With(slog.String("op", op))

	b, err := json.Marshal(e)
	if err != nil {
		log.Error("unable to marshal event", slog.String("err", err.Error()))
		return
	}

	ak.asyncProducer.Input() <- &sarama.ProducerMessage{
		Topic: ak.topic,
		Value: sarama.ByteEncoder(b),
	}
}

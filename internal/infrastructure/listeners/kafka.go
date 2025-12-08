package listeners

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
)

type Kafka struct {
	consumer sarama.Consumer
	topic    string
}

func NewKafka(consumer sarama.Consumer, topic string) *Kafka {
	return &Kafka{
		consumer: consumer,
		topic:    topic,
	}
}

func (k *Kafka) Start(ctx context.Context) (<-chan []byte, error) {
	const op = "listeners.KafkaListener.Start"

	log := slog.With(slog.String("op", op))

	partitions, err := k.consumer.Partitions(k.topic)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ch := make(chan []byte, len(partitions))

	wg := new(sync.WaitGroup)
	for _, partition := range partitions {
		partitionConsumer, err := k.consumer.ConsumePartition(k.topic, partition, sarama.OffsetOldest)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-partitionConsumer.Messages():
					ch <- msg.Value
				case err := <-partitionConsumer.Errors():
					log.Error("consumer error", slog.String("err", err.Error()))
					return
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch, nil
}

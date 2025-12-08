package configs

type Poster struct {
	KafkaHosts         []string `envconfig:"KAFKA_HOSTS"`
	PublishedPostTopic string   `envconfig:"PUBLISHED_POST_TOPIC"`
	Database           Postgres
}

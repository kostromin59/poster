package configs

type Poster struct {
	KafkaHosts         []string `envconfig:"KAFKA_HOSTS" required:"true"`
	PublishedPostTopic string   `envconfig:"PUBLISHED_POST_TOPIC" required:"true"`
	TGBotToken         string   `envconfig:"TG_BOT_TOKEN" required:"true"`
	Location           string   `envconfig:"LOCATION" default:"Asia/Yekaterinburg"`
	TGPublishChatID    int64    `envconfig:"TG_PUBLUSH_CHAT_ID" required:"true"`
	TGAllowedUsers     []int64  `envconfig:"TG_ALLOWED_USERS" required:"true"`
	Database           Postgres
	Redis              Redis
}

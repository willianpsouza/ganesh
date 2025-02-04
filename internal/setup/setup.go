package setup

import "time"

const (
	NatsAddress     string        = "nats://localhost:4222"
	QueueName       string        = "ganesh.provengo.io"
	TotalTasks      int           = 2
	QueueGroup      string        = "login_workers"
	ConsumerType    string        = "workers"
	UsersGenerate   int           = 1000
	RedisDefaultTTL time.Duration = 120 * time.Minute
	PostgresURI     string        = "postgres://postgres:postgres123@localhost:5432/postgres?sslmode=disable"
	PostgresMin     int32         = 16
	PostgresMax     int32         = 64
)

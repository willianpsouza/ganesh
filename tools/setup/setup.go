package setup

import "time"

const (
	QueueName              string        = "ganesh.provengo.io"
	TotalTasks             int           = 2
	QueueGroup             string        = "login_workers"
	ConsumerType           string        = "workers"
	UsersGenerate          int           = 100000
	RedisDefaultTTL        time.Duration = 120 * time.Minute
	PostgresURI            string        = "postgres://postgres:meni4na6@localhost:5432/postgres?sslmode=disable"
	PostgresDB             string        = "postgres"
	PostgresUser           string        = "postgres"
	PostgresPassword       string        = "meni4na6"
	PostgresTransactionTTL time.Duration = 120 * time.Minute
	PostgresMin            int32         = 16
	PostgresMax            int32         = 64
)

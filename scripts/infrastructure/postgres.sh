#!/bin/sh

if [ -z "$PG_PASSWORD" ]; then
	echo "Defina a variavel export PG_PASSWORD=<senha>"
	exit 1
fi

sudo docker run --name local_postgres --ulimit nofile=65536:65536 --memory=4096m --cpus=4 --cpuset-cpus="8,9,10,11" --rm -p 5432:5432 -e POSTGRES_PASSWORD=${PG_PASSWORD} -d -v postgres_data:/var/lib/postgresql/data postgres


<<'COMMENT'
sudo docker exec -ti local_postgres psql -U postgres
CREATE TABLE users (
    uuid text PRIMARY KEY,
    username text,
    password text,
    hash text
);
INSERT INTO tasks (description) VALUES ('Finish work'), ('Have fun');
COMMENT

#!/bin/sh



#-sudo docker run --name local_nats --network nats --rm -p 4222:4222 -p 8222:8222 nats --http_port 8222
sudo docker run --name local_postgres --cpus=2 --rm -p 5432:5432 -e POSTGRES_PASSWORD=${PASSWORD} -d -v postgres_data:/var/lib/postgresql/data postgres

sudo docker exec -ti local_postgres psql -U postgres


```
CREATE TABLE users (
    uuid text PRIMARY KEY,
    username text,
    password text,
    hash text
);
INSERT INTO tasks (description) VALUES ('Finish work'), ('Have fun');
```
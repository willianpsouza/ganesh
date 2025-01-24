#!/bin/sh

#-sudo docker run --name local_nats --network nats --rm -p 4222:4222 -p 8222:8222 nats --http_port 8222
sudo docker run --name local_nats --cpus=1 --rm -p 4222:4222 -p 8222:8222 nats --http_port 8222

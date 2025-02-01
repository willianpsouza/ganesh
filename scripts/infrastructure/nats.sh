#!/bin/sh

#-sudo docker run --name local_nats --network nats --rm -p 4222:4222 -p 8222:8222 nats --http_port 8222
sudo docker run --name local_nats --ulimit nofile=65536:65536 --memory=512m --cpus=2 --cpuset-cpus="14,15" --rm -p 4222:4222 -p 8222:8222 nats --http_port 8222

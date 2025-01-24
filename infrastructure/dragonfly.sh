#!/bin/sh

#docker pull docker.dragonflydb.io/dragonflydb/dragonfly
#sudo docker run --name local_dragonfly --cpus=1 --rm -p 6379:6379 docker.dragonflydb.io/dragonflydb/dragonfly

sudo docker run --name local_dragonfly --ulimit nofile=65536:65536 --cpus=2 --rm -p 6379:6379 docker.dragonflydb.io/dragonflydb/dragonfly --maxmemory=2gb --dbfilename=""

#!/bin/sh

export ALGORITHM=BCRYPT
export GIN_MODE=release
export GOMAXPROCS=2


bin/./main

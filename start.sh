#!/bin/bash

docker build -t zipper .
docker run -ti -p 9191:80 zipper
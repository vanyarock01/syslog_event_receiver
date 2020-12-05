#!/bin/bash

docker run -d \
           --name syslog-storage \
           -p 3301:3301 \
           -v $(pwd)/storage/:/opt/tarantool/ \
           tarantool/tarantool:2.6 \
           tarantool /opt/tarantool/init.lua
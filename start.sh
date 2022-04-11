#!/bin/bash

while true; do
    LD_LIBRARY_PATH="/usr/local/lib/:$LD_LIBRARY_PATH" \
    /root/extend-api-service -r /root 2>&1 | tee -a /var/log/extend-api.log > /dev/null
done

#!/bin/sh
sleep 1
/etc/init.d/cron start
service cron start
/root/start.sh 

FROM ubuntu

RUN apt-get update && \
	apt-get -y upgrade

RUN	apt-get install -y apt-utils && \
  apt-get install -y tzdata && \
  cp -p /usr/share/zoneinfo/US/Pacific /etc/localtime && \
	apt-get install -y unzip && \
    apt-get install -y vim curl cron logrotate && \
    apt-get install nginx python3-certbot-nginx mc git wget -y &&\
    apt-get clean all

# Set the working directory to /root
WORKDIR /root

RUN mkdir -p /root/.aws
# RUN (crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | crontab -
RUN (crontab -l 2>/dev/null; echo "0 12 * * * logrotate /etc/logrotate.d/extend-api") | crontab -
ADD startup.sh /root
ADD start.sh  /root/
ADD extend-api-service /root/
ADD extend-api       /etc/logrotate.d/extend-api
ADD version /root/
EXPOSE 8000

ENTRYPOINT /root/startup.sh; /bin/bash

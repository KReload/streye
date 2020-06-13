FROM node:10.15.2

WORKDIR /usr/src/app
COPY . .
RUN npm install --save
RUN npm run build

FROM node:10.15.2

WORKDIR /usr/src/app
COPY package*.json ./
RUN npm install --only=production


FROM tiangolo/nginx-rtmp

RUN apt-get update && apt-get install -y \
    apache2-utils

RUN htpasswd -b -c /etc/nginx/.htpasswd streye eyerts &&\
    mkdir /tmp/hls &&\
    mkdir /tmp/rec &&\
    chown -R nobody:root /tmp/rec &&\
    chmod -R 775 /tmp/rec

COPY --from=0 /usr/src/app/public /data/www/

EXPOSE 80 1935
#! /bin/sh

rm -rf /usr/share/nginx/html && \
tar xfz /tmp/html.tar.gz -C /usr/share/nginx/ && \
nginx "-g daemon off;"
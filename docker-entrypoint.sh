#!/bin/sh
# Start Go backend on internal port
export PORT=3031
/app/wa-assistant &

# Start nginx in foreground
nginx -g 'daemon off;'

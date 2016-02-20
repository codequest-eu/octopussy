#!/bin/sh

trap 'kill -9 $(jobs -p)' EXIT
echo -e $WS_CERT > cert.pem
echo -e $WS_KEY > cert.key
$@

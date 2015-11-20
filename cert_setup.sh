#!/bin/bash
echo $WS_CERT > cert.pem
echo $WS_KEY > cert.key
$@

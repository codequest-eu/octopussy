#!/bin/bash
echo -e $WS_CERT > cert.pem
echo -e $WS_KEY > cert.key
$@

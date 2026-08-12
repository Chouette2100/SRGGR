#!/bin/sh
filename=`date +%Y%m%d-%H%M`
tar zcvf SRGGR_$filename.tar.gz \
DBConfig*.enc.yaml \
Env.yml \
SRGGR \
srggr.sh \
tar.sh

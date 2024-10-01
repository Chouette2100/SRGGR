#!/bin/bash
filename=`date +%Y%m%d-%H%M`
tar zcvf SRGGR_$filename.tar.gz \
DBConfig.yml \
Env.yml \
SRGGR \
srggr.sh

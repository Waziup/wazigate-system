#!/bin/bash

loragna_g=$(python /app/conf.py cell_conf loragna_g)
#echo $loragna_g;

if [[ "$loragna_g" == "3G" ]];
then
	python wake-3G.py
else	
	python wake-2G.py
fi
	
sleep 2
(
    while : ; do
        pppd $1 call gprs
        sleep 10
    done
) &

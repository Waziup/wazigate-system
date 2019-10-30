#!/bin/bash
chmod +x /app/data_acq/lora/lora_gateway

#Check if all requirements are installed. which usually have problems in installation during the image building
#if ! [ -x "$(which iw)" ]; then  fi;
apk add iw
apk add gawk
apk add nano


python /app/setup.py

python /app/api.py & 
python /app/startLora.py

#python /app/api.py

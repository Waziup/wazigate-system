#!/bin/bash
chmod +x /app/data_acq/lora/lora_gateway

python /app/setup.py

python /app/api.py & 

noLora=$NO_LORA
if [ "$noLora" = true ]; then
      echo "Lora not started"
else
      python /app/startLora.py &
fi

#Waiting for Ctrl-C
while :; do sleep 1; done


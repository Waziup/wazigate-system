#!/bin/bash
chmod +x /app/data_acq/lora/lora_gateway

python -u /app/setup.py

python -u /app/api.py & 

noLora=$NO_LORA
if [ "$noLora" = true ]; then
      echo "Lora not started"
else
      python -u /app/startLora.py &
fi

#Waiting for Ctrl-C
while :; do sleep 1; done


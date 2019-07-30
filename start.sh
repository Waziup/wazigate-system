#!/bin/bash
chmod +x /app/data_acq/lora/lora_gateway

python /app/setup.py

python /app/api.py &
python /app/startLora.py

#Image for compiling
FROM python:alpine as compile

RUN apk update && \
    apk add python-dev zlib-dev jpeg-dev linux-headers gcc g++ make libffi-dev openssl-dev build-base networkmanager wpa_supplicant grep libc6-compat
    # wvdial gammu python-gammu


#installing Python packages
WORKDIR /app
COPY requirements.txt /app
RUN pip install --user -r requirements.txt

#compiling lora_gateway C executable
WORKDIR /app/data_acq/lora
COPY data_acq/lora /app/data_acq/lora
RUN make lora_gateway_pi2

#Minimal image for execution
FROM python:alpine as run

RUN apk update && \
    apk add nano gammu iw gawk wpa_supplicant
    # wvdial

#Copy build results
COPY --from=compile /root/.local /root/.local
COPY --from=compile /app/data_acq/lora/lora_gateway /app/data_acq/lora/lora_gateway 
WORKDIR /app
COPY . /app

# Let's leave the 3G support to the next version
#RUN wget https://github.com/wlach/wvdial/archive/master.zip
#RUN unzip master.zip && \
#	cd wvdial-master && \
#	make && make install

RUN chmod +x start.sh
ENTRYPOINT [ "sh", "start.sh" ]
FROM python:alpine as compile

MAINTAINER Moji eskandari@fbk.eu

RUN apk update && \
    apk add python-dev zlib-dev jpeg-dev linux-headers gcc g++ make libffi-dev openssl-dev 
    # wvdial gammu python-gammu

WORKDIR /app
COPY requirements.txt /app
RUN pip install --user -r requirements.txt


FROM python:alpine as run

RUN apk update && \
    apk add iw gawk networkmanager nano wpa_supplicant 


WORKDIR /app
COPY --from=compile /root/.local /root/.local
COPY . /app

RUN chmod +x ./start.sh
ENTRYPOINT [ "sh", "./start.sh" ]

FROM python:latest

MAINTAINER Moji eskandari@fbk.eu

RUN apt-get update -y && \
    apt-get install -y python-pip python-dev iw gawk network-manager nano wvdial gammu python-gammu
   
RUN pip install flask wifi requests pyserial 

COPY . /app
WORKDIR /app/

RUN chmod +x ./start.sh
ENTRYPOINT [ "sh", "./start.sh" ]

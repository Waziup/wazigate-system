#-------------------------------------------------------------------------------
# @author: Moji on June 06th 2019 to call the WaziEdge
# inspired by a code from Congduc Pham
# eskandari@fbk.eu

try:
	from urllib.request import urlopen
	from urllib.request import URLError
except ImportError:
	from urllib2 import urlopen
	from urllib2 import URLError

import requests
#import subprocess
import time
#import ssl
#import socket
import datetime
import sys
import os
import json
import re
from hashlib import md5
#import shlex

#import sys
sys.path.insert( 0, os.path.dirname( os.path.dirname( os.path.abspath(__file__))));
import api

conf = api.conf_read();

#-----------------------------------------#

EdgeHeaders = {'accept':'application/json','content-type':'application/json'}

addr = os.environ['WAZIGATE_EDGE_ADDR'].split(':');
EdgeURL = "http://localhost";

if( len( addr) > 0 and len( addr[0]) > 0):
	EdgeURL = addr[0];

if( len( addr) > 1 and len( addr[1]) > 0):
	EdgeURL += ':' + addr[1];

#-----------------------------------------#

def sendToEdge( devId, sensorId, value):

	if( len( devId) == 0):
		#If there is no Device ID use the hash of the Gateway ID instead
		devId = md5( conf['gateway_conf']['gateway_ID'].encode("utf-8")).hexdigest();

	sensorURL	= EdgeURL +'/devices/'+ devId +'/sensors/'+ sensorId +'/value';
	sensorValue	= json.dumps( value);

	try:
		response = requests.post( sensorURL, headers = EdgeHeaders, data = sensorValue, timeout = 30);
		#print( response.url, response.status_code);

		if( response.status_code == 404): #The device or the sensor do not exist
			
			url = EdgeURL +'/devices/'+ devId;
			response = requests.get( url, headers = EdgeHeaders, timeout = 30);
			
			if( response.status_code == 404): #The device does not exist, Creating it
				url = EdgeURL +'/devices/';
				newDeviceData = json.dumps( { "id": devId, "name": devId});
				response = requests.post( url, headers = EdgeHeaders, data = newDeviceData, timeout = 30);
				if( response.ok):
					print( 'New device created with ID: ', devId);
				else:
					print( 'Error! Failed creating new device with ID: ', devId);
					print( response.url);
					print( response.content);

			#---------#
			
			url = EdgeURL +'/devices/'+ devId +'/sensors/'+ sensorId;
			response = requests.get( url, headers = EdgeHeaders, timeout = 30);
			#print( response.url, response.status_code);

			if( response.status_code == 404): #The Sensor does not exist, Creating it
				url = EdgeURL +'/devices/'+ devId +'/sensors/';
				newSensorData = json.dumps( { "id": sensorId, "name": sensorId});
				response = requests.post( url, headers = EdgeHeaders, data = newSensorData, timeout = 30);
				if( response.ok):
					print( 'New sensor created with ID: ', sensorId);
				else:
					print( 'Error! Failed creating new sensor with ID: ', sensorId);
					print( response.url);
					print( response.content);
			
			#---------#
			
			response = requests.post( sensorURL, headers = EdgeHeaders, data = sensorValue, timeout = 30);
			#print( response.url, response.status_code);
			
		if response.ok:
			print( 'Edge: upload success', sensorId, value);
#			print( response.url, response.status_code);
		else:
			print( 'Edge: bad request');
			print( response.url, response.status_code);
			
	except requests.exceptions.RequestException as e:
		print(e);

	return 0;

#-----------------------------------------#

if __name__ == "__main__":
	ldata	= sys.argv[1];
	pdata	= sys.argv[2];
	#rdata	= sys.argv[3];
	#tdata	= sys.argv[4];

	# this is common code to process packet information provided by the main gateway script (i.e. post_processing_gw.py)
	# these information are provided in case you need them
	arr	= list( map( int, pdata.split(',')));
	dst	= arr[0]
#	ptype=arr[1]				
	src=arr[2]
#	seq=arr[3]
#	datalen=arr[4]
#	SNR=arr[5]
#	RSSI=arr[6]

	#LoRaWAN packet
	if dst == 256:
		src_str = "%0.8X" % src
	else:
		src_str = str(src)	

	#remove any space in the message as we use '/' as the delimiter
	#any space characters will introduce error in the json structure and then the curl command will fail
	ldata = ldata.replace( ' ', '');
	data_array = re.split( "/", ldata);

	# just in case we have an ending CR or 0
	data_array[ len( data_array) - 1 ] = data_array[ len( data_array) - 1].replace( '\n', '');
	data_array[ len( data_array) - 1 ] = data_array[ len( data_array) - 1].replace( '\0', '');
	
	data = { data_array[i]: data_array[i+1] for i in range(0, len( data_array), 2)};
	
	print( 'Received data (from '+ src_str +'): ', data);
	
	#Check for the device ID
	devId = '';
	if( 'UID' in data):
		devId = data['UID'];
		del data['UID'];

	for sensorId in data:
		sendToEdge( devId, sensorId, data[ sensorId]);


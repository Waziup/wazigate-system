#-------------------------------------------------------------------------------
# Copyright 2018 Congduc Pham, University of Pau, France.
# 
# Congduc.Pham@univ-pau.fr
#
# This file is part of the low-cost LoRa gateway developped at University of Pau
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
# 
# You should have received a copy of the GNU General Public License
# along with the program.  If not, see <http://www.gnu.org/licenses/>.
#-------------------------------------------------------------------------------

# Modified by Moji on April 30th 2019 to call the WaziEdge
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
edge_conf = conf['edge_conf'];

# didn't get a response from server?
connection_failure = False

#the token
current_token="notoken"
last_token_time="notime"
token_valid_time_in_seconds=2*60

#common header for requests
Edge_headers = {'accept':'application/json','content-type':'application/json'}

gw_id_md5 = ''

global CloudNoInternet_enabled

num_format = re.compile( "^[\-]?[1-9][0-9]*\.?[0-9]+$");


# function to check connection availability with the server
def test_network_available():
	connection = False
	iteration = 0
	response = None
	
	# we try 4 times to connect to the server.
	while(not connection and iteration < 4) :
		try:
			# 3sec timeout in case of server available but overcrowded
			response = urlopen( edge_conf['server'] +'/sensors', timeout=3)
			connection = True
		except URLError as e: pass
		except socket.timeout: pass
		except ssl.SSLError: pass
	    	
		# if connection_failure == True and the connection with the server is unavailable, don't waste more time, exit directly
		if(connection_failure and response is None) :
			print('Edge: the server is still unavailable');
			iteration = 4
			# print(connection failure);
		elif(response is None) :
			print('Edge: server unavailable, retrying to connect soon...');
			# wait before retrying
			time.sleep(1)
			iteration += 1
	    		
	return connection

#implemented policy
# - username is guest, regardless of visibility: can push data (public) but entity name is appended with gw id md5 hash
# - username is not guest, regardless of visibility: if authentification failed, then abort
		
def get_token():

	global connection_failure
	global current_token
	global last_token_time
	global token_valid_time_in_seconds
	global Edge_headers

	if len( edge_conf['username']) == 0 or edge_conf['username'] == 'guest':
		if current_token != "guest":
			print( "CloudEdge: run under guest identity, all data is public");
			current_token = "guest"
		Edge_headers = {'accept':'application/json','content-type':'application/json'};
		return current_token;
				
	ask_new_token=False
	
	if current_token=="notoken":
		ask_new_token=True
	else:
		token_elapsed_time=datetime.datetime.now()-last_token_time		
		if token_elapsed_time.total_seconds() > token_valid_time_in_seconds:
			ask_new_token=True
	 		
	if ask_new_token:
			
		Edge_url = edge_conf['server'] +'/auth/token';
		my_Edge_headers = {'accept':'application/json','content-type':'application/json'};
		Edge_data = '{"username":"'+ edge_conf['username'] +'","password":"'+ edge_conf['password'] +'"}';
	
		print( "CloudEdge: get token for "+ edge_conf['username']);
		print( "CloudEdge: will issue requests with");
		print( 'url: '+ Edge_url);
		#print('headers: '+json.dumps(Edge_headers)
		#print('data: '+Edge_data		
	
		current_token="notoken"
	
		try:
			response = requests.post(Edge_url, headers=my_Edge_headers, data=Edge_data, timeout=30)

			print('CloudEdge: returned msg from server is ', end='');
			print(response.status_code);
						
			if response.status_code == 200:
				print('CloudEdge: got token');
				current_token=response.text
				Edge_headers = {'accept':'application/json','content-type':'application/json','Authorization':'Bearer '+current_token}
				last_token_time=datetime.datetime.now()
			else:	
			#if response.status_code == 404:
				print('CloudEdge: authorization failed');
				
		except requests.exceptions.RequestException as e:
			print(e);
			print('CloudEdge: requests command failed (maybe a disconnection)');
			connection_failure = True		
		
	return current_token	
	
def create_new_measurement(entity, nomenclature):

	global connection_failure
	global gw_id_md5
	
	my_token=get_token()
	
	if my_token=="notoken":
		return False
		
	Edge_url= edge_conf['server'] +'/sensors/'+ entity +'/measurements';
	Edge_data = '{"id":"'+ nomenclature +'"}';

	print('CloudEdge: create new measurement');
	print("CloudEdge: will issue requests with");
	print('url: '+Edge_url);
	#print('headers: '+json.dumps(Edge_headers);
	print('data: '+Edge_data);

	try:
		response = requests.post(Edge_url, headers=Edge_headers, data=Edge_data, timeout=30)
	
		print('CloudEdge: returned msg from server is ',end='');
		print(response.status_code);
			
		if response.status_code==200:
			print('CloudEdge: measurement creation success');
			return True			
		else:
			print('CloudEdge: error in creating measurement');
			return False
					
	except requests.exceptions.RequestException as e:
		print(e);
		print('CloudEdge: requests command failed (maybe a disconnection)');
		connection_failure = True	
	
	return False

#check if measurement exists
def does_measurement_exist(entity, nomenclature):

	global connection_failure

	my_token=get_token()
	
	if my_token=="notoken":
		return False

	Edge_url= edge_conf['server'] +'/sensors';
	Edge_url= Edge_url +'/'+ entity		
	Edge_url= Edge_url +'/measurements/'+ nomenclature

	print('CloudEdge: check if measurement exits');
	print("CloudEdge: will issue requests with");
	print('url: '+Edge_url);
	#print('headers: '+json.dumps(Edge_headers);

	try:
		response = requests.get(Edge_url, headers=Edge_headers, data='', timeout=30)

		print('CloudEdge: returned msg from server is ',end='');
		print(response.status_code);
					
		if response.status_code==404:
			print('CloudEdge: measurement does not exist');
			return False
		elif response.status_code==200:
			print('CloudEdge: measurement exists');
			return True				
		else:
			print('CloudEdge: error in checking measurement');
			return False
			
	except requests.exceptions.RequestException as e:
		print(e);
		print('CloudEdge: requests command failed (maybe a disconnection)');
		connection_failure = True
	
	if connection_failure:
		return False		

def does_entity_exist(entity):

	global connection_failure

	my_token=get_token()
	
	if my_token=="notoken":
		return False
		
	#check if entity exist
	Edge_url= edge_conf['server'] +'/sensors'
	Edge_url=Edge_url+'/'+entity
	
	print('CloudEdge: check if entity exits');
	print("CloudEdge: will issue requests with");
	print('url: '+Edge_url);
	#print('headers: '+json.dumps(Edge_headers)
	
	try:
		response = requests.get(Edge_url, headers=Edge_headers, data='', timeout=30)

		print('CloudEdge: returned msg from server is ',end='');
		print(response.status_code);
						
		if response.status_code==404:
			print('CloudEdge: entity does not exist');
			return False
		elif response.status_code==200:
			print('CloudEdge: entity exists');
			return True
		else:
			print('CloudEdge: error in checking entity');
			return False							

	except requests.exceptions.RequestException as e:
		print(e);
		print('CloudEdge: requests command failed (maybe a disconnection)');
		connection_failure = True
		
	if connection_failure:
		return False
		
def create_new_entity(data, entity, nomenclatures, tdata):

	global connection_failure
	global gw_id_md5
			
	my_token = get_token()
	
	if my_token == "notoken":
		return False		
		
	Edge_url	= edge_conf['server']+'/sensors' 
	Edge_data	= '{"id":"'+ entity +'","gateway_id":"'+ gw_id_md5 +'","domain":"'+ edge_conf['domain'] +'"';
	
	Edge_data = Edge_data +',"visibility":"private"';
	#Edge_data = Edge_data+',"visibility":"public"';
	
	Edge_data = Edge_data +'}'
	
	print( "CloudEdge: create new entity");
	print( "CloudEdge: will issue requests with");
	print( 'url: '+ Edge_url);
	#print('headers: '+json.dumps(Edge_headers);
	print( 'data: '+ Edge_data);
	
	try:
		response = requests.post(Edge_url, headers=Edge_headers, data=Edge_data, timeout=30)

		print( 'CloudEdge: returned msg from server is ', end='');
		print( response.status_code);
						
		if response.ok:
			print( 'CloudEdge: entity creation success');
			print( "CloudEdge: create measurements for new entity");
		
			i=0
	
			while i < len( nomenclatures):
				create_new_measurement(entity, nomenclatures[i])
				#create next measurement in same entity				
				i += 1			
		else:
			print( 'CloudEdge: error in creating entity');
			return False
				
	except requests.exceptions.RequestException as e:
		print( e);
		print( 'CloudEdge: requests command failed (maybe a disconnection)');
		connection_failure = True
	
	if connection_failure:
		return False
	else:
		return True	
	 	
		
# send a data to the server
def send_data(data, entity, nomenclatures, tdata):

	global connection_failure
	
	if data[0]=='':
		data[0]=edge_conf['domain']

	if data[1]=='':
		data[1]=edge_conf['domain']

	#we ask for a token right now
	my_token=get_token()
	
	if my_token=="notoken":
		return False
			
	#check entity
	if does_entity_exist(entity)==False:
		if create_new_entity(data, entity, nomenclatures, tdata)==False:
			return False
			
	i=0
	
	#will try to upload values for each measurement			
	while i < len(data)-2:

		#check measurement
		if does_measurement_exist(entity, nomenclatures[i])==False:
			if create_new_measurement(entity, nomenclatures[i])==False:
				return False
				
		Edge_url	= edge_conf['server'] +'/sensors/'+ entity +'/measurements/'+ nomenclatures[i] +'/values';
		Edge_data	= '{"value":'

		isnumber = re.match(num_format,data[i+2])
		#isnumber = False
				
		if isnumber:
			Edge_data = Edge_data+data[i+2]+',"timestamp":"'+tdata+'"}'
		else:
			Edge_data = Edge_data+'"'+data[i+2]+'","timestamp":"'+tdata+'"}'
					
		print("CloudEdge: will issue requests with");
		print('url: '+Edge_url);
		#print('headers: '+json.dumps(Edge_headers)
		print('data: '+Edge_data);
		
		try:
			response = requests.post(Edge_url, headers=Edge_headers, data=Edge_data, timeout=30)

			print('CloudEdge: returned msg from server is ',end='');
			print(response.status_code);

			if response.ok:
				print('CloudEdge: upload success');
				i += 1
			else:
				print('CloudEdge: bad request');
				i += 1
				
		except requests.exceptions.RequestException as e:
			print(e);
			print('CloudEdge: requests command failed (maybe a disconnection)');
			connection_failure = True	
			
		if connection_failure:
			return False			

# main
# -------------------
#
# ldata can be formatted to indicate a specific project_name and service_tree. Options are:
# 	TC/22.4/HU/85 -> use default project_name and service_tree
#	-TESTS#TC/22.4/HU/85 -> use default project_name and service_tree=-TESTS
#	waziup#-TEST#TC/22.4/HU/85 -> project_name=waziup and service_tree=-TESTS
#
#	project_name and service_tree must BOTH have more than 2 characters
#
#   the domain will be project_name+'-'+organization_name+service_tree, e.g. waziup-UPPA-TESTS
#	the entity name will be organization_name+service_tree+'_'+sensor_name+scr_addr, e.g UPPA-TESTS_Sensor2
#
#	you can test the script in standalone mode as follwow:
#
#	python CloudEdge.py "TC/22.5" "1,16,36,0,9,8,-45" "125,5,12" "2018-08-05T11:08:52" "00000027EB5A71F7"
#
def main(ldata, pdata, rdata, tdata, gwid):

	# this is common code to process packet information provided by the main gateway script (i.e. post_processing_gw.py)
	# these information are provided in case you need them
	arr = list( map( int, pdata.split(',')));
	dst=arr[0]
	ptype=arr[1]				
	src=arr[2]
	seq=arr[3]
	datalen=arr[4]
	SNR=arr[5]
	RSSI=arr[6]
	
	#compute the MD5 digest (hash) from the clear gw id provided by post_processing_gw.py
	global gw_id_md5
	#gw_id_md5=md5.new(gwid).hexdigest()
	gw_id_md5=md5( gwid.encode("utf-8")).hexdigest()
	
	
	#LoRaWAN packet
	if dst==256:
		src_str="%0.8X" % src
	else:
		src_str=str(src)	

	if True: #( src_str in key_Edge.source_list) or (len(key_Edge.source_list)==0):

		#remove any space in the message as we use '/' as the delimiter
		#any space characters will introduce error in the json structure and then the curl command will fail
		ldata=ldata.replace(' ', '')
			
		# this part depends on the syntax used by the end-device
		# we use: TC/22.4/HU/85...
		#
		# but we accept also a_str#b_str#TC/22.4/HU/85... to indicate a project_name and a service_path
		# or simply 22.4 in which case, the nomemclature will be DEF
		 		
		# get number of '#' separator
		nsharp=ldata.count('#')
		nslash=0
				
		# no separator
		if nsharp==0:
			# will use default project_name and service_path
			data=['','']

			# get number of '/' separator on ldata
			nslash = ldata.count('/')
				
			# contains ['', '', "s1", s1value, "s2", s2value, ...]
			data_array = data + re.split("/", ldata)		
		else:
			data_array = re.split("#", ldata)
		
			# only 1 separator
			if nsharp==1:
				# insert '' to indicate default project name
				# as we assume that the only parameter indicate the service_path
				data_array.insert(0,'');
				# if the length is greater than 2
				if len(data_array[1])<3:
					data_array[1]=''	

			# we have 2 separators
			if nsharp==2:
				# if the length of a fields is greater than 2 then we take it into account
				if len(data_array[0])<3: 
					data_array[0]=''				
				if len(data_array[1])<3:
					data_array[1]=''
									
			# get number of '/' separator on data_array[2]
			# because ldata may contain '/' in service_path name
			nslash = data_array[2].count('/')
	
			# then reconstruct data_array
			data_array=[data_array[0],data_array[1]]+re.split("/", data_array[2])
				
			# at the end data_array contains
			# ["project_name", "service_path", "s1", s1value, "s2", s2value, ...]
		
		# just in case we have an ending CR or 0
		data_array[len(data_array)-1] = data_array[len(data_array)-1].replace('\n', '')
		data_array[len(data_array)-1] = data_array[len(data_array)-1].replace('\0', '')	
																		
		nomenclatures = []
		# data to send
		data = []
		data.append(data_array[0]) #project_name (if '' default)
		data.append(data_array[1]) #service_path (if '' default)
		
		if nslash==0:
			# old syntax without nomenclature key, so insert only one key
			# we use DEF
			nomenclatures.append("DEF")
			data.append(data_array[2])
		else:
			# completing nomenclatures and data
			i=2
			while i < len(data_array)-1 :
				nomenclatures.append(data_array[i])
				data.append(data_array[i+1])
				i += 2

		connected = test_network_available()
	
		#if we got a response from the server, send the data to it	
		if (connected):
			print("CloudEdge: uploading");
			#here we append the device's address to get for instance Sensor2
			#if packet come from a LoRaWAN device with 4-byte devAddr then we will have for instance Sensor01020304
			#where the devAddr is expressed in hex format
			
			append_gw_str=''
	
			#default organization name or default username?
			if edge_conf['username'] == '' or edge_conf['username'] == 'guest':
				#we append the md5 hash to the sensor name: ORG_Sensor2_4b13a223f24d3dba5403c2727fa92e62
				append_gw_str='_'+gw_id_md5	
				
			#key_Edge.service_path=key_Edge.organization_name+key_Edge.service_tree
			#the provided parameter would be the entity name, i.e. UPPA-TESTS_Sensor2
			#here: UPPA is the organization_name, -TESTS is the servire_tree
			
			#if organization name is ORG (default) or username is guest (default) then we append the gw id md5 hash
			
			send_data(data, edge_conf['domain'] +'_'+ edge_conf['sensor']+src_str+append_gw_str, nomenclatures, tdata);

		else:
			print("CloudEdge: not uploading");
			
			if (CloudNoInternet_enabled):
				print("Using CloudNoInternet");
				from CloudNoInternet import store_internet_pending
				# we call store_internet_pending to store the message for future upload
				store_internet_pending(ldata, pdata, rdata, tdata, gwid)

		# update connection_failure value
		global connection_failure
		connection_failure = not connected

	else:
		print("Source is not is source list, not sending with edgeCall.py");

if __name__ == "__main__":
	main( sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5])

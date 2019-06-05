import sys
import json
import os
import api

#------------------------#

conf = api.conf_read();

try:
	if( conf['setup_conf']['gwid_done'] == True):
		quit();
except KeyError as e:
	print( "Error: Key error on setup conf!");
	print( e);
	quit();
	pass;

#---------- Setting up the Gateway ID and Access point -------------#

#TODO: Config the gateway hotspot as well.

jres = api.gwid();
gwid = json.loads( jres[0]);

ap_conf = {
	'SSID'		:	'WAZIGATE_'+ gwid,
	'password'	:	'loragateway',
	'interface'	:	api.WIFI_DONGLE
};

print( api.wifi_set_ap( ap_conf));


newConf = {
	'setup_conf': {'gwid_done': True}
	};

api.conf_set( newConf);

#--------------------------------------------------------------------#

#Reboot now...
print( api.system_shutdown( 'reboot'));

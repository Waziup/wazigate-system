import sys
import json
import os
import api
from shutil import copyfile

#------------------------#

#Check if the conf.json file exist
if( os.path.isfile( api.CONF_FILE) == False):
	#Copying from default file
	copyfile( api.PATH +'/conf/conf.default.json', api.CONF_FILE);

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

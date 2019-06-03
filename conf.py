import sys
import json
import os
import api

#------------------------#

if len( sys.argv) > 2: # We need at least two parameters to get the config 
	conf = api.conf_read();

	if( sys.argv[1] in conf):
		if( sys.argv[2] in conf[sys.argv[1]]):
			print( str( conf[ sys.argv[1] ][ sys.argv[2] ]).strip(), end='');

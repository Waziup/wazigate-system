#!/usr/bin/python
from flask import Flask, send_from_directory
from flask import request
import subprocess
import json
import ast
import time
import os

#------------------------#

API_VER		=	"v1";

#Path to the root of framework
PATH		=	os.path.dirname(os.path.abspath(__file__));
SCRIPTS		=	PATH + '/scripts/';
CONF_FILE	=	PATH + '/conf.json'; 
LOGS_PATH	=	PATH + '/logs/';

WIFI_DEVICE = 'wlan0'; #'wlp2s0'; # The wifi device which connects to the internet. e.g. wlan0
WIFI_DONGLE = 'wlan1';
ETH_DEVICE  = 'eth0';

if( 'WIFI_DEVICE' in os.environ):
	WIFI_DEVICE = os.environ['WIFI_DEVICE'];

if( 'WIFI_DONGLE' in os.environ):
	WIFI_DONGLE = os.environ['WIFI_DONGLE'];
	
if( 'ETH_DEVICE' in os.environ):
	ETH_DEVICE = os.environ['ETH_DEVICE'];


#------------------------#

app = Flask(__name__);
#app = Flask(__name__, static_url_path = PATH + '/docs/');

#------------------------#

@app.route('/')
@app.route('/<path:filename>')
def index( filename = ''):
	if( len( filename) == 0):
		filename = 'index.html';
	return send_from_directory( PATH + '/docs/', filename);
	return "Salam Goloooo!"

#------------------------#

@app.route('/shell', methods=['POST'])
def shell():
	if not request.json:
		return 'Not JSON input', 500;
	res = cmd; #os.popen( request.json['cmd']).read();
	return json.dumps( res), 201;

#------------------------#

def evalfn( pairs):
	res = {}
	for key, val in pairs:
		if val in {'true','false'}:
			res[key] = val == 'true'
			continue
		try:
			res[key] = ast.literal_eval(val)
		except Exception as e:
			res[key] = val
	return res

#----------------#

def evalVal( val):
	if type( val) is list:
		return val;
	if val in {'true','false'}:
		return val == 'true';
	try:
		return ast.literal_eval( val);
	except Exception as e:
		return val;

#------------------------#

def conf_read():
	res = [];
	with open( CONF_FILE) as f:
		res = json.load( f);
	return res;

#------------------------#

@app.route('/api/'+ API_VER +'/system/conf', methods=['GET'])
def conf_get():
	return json.dumps( conf_read()), 201;

#------------------------#

@app.route('/api/'+ API_VER +'/system/conf',  methods=['PUT', 'POST'])
def conf_set_api():
	if( request.json is None):
		return "JSON Err!", 201;

	#return json.dumps( request.json), 201

	if( 'json' not in request.json):
		return "Input Err!", 201;

	conf_set( request.json['json']);

	return "Ok", 201;

#------------------------#

def conf_set( cfg):
	with open( CONF_FILE, 'r+') as f:
		data = json.load(f);
		
		for key in cfg:
			for sub in cfg[ key ]:
				data[ key ][ sub ] = evalVal( cfg[ key ][ sub ]);
		f.seek( 0);
		f.write( json.dumps(data, indent=2));
		f.truncate();
	return 0;

#------------------------#

@app.route('/api/'+ API_VER +'/system/net', methods=['GET'])
def net_get():
	
	cmd = 'ip route show default | head -n 1 | awk \'/default/ {print $5}\'';
	dev = os.popen( cmd).read().strip();
	
	if( len( dev) == 0):
		return "", 201;
	
	cmd = 'cat /sys/class/net/'+ dev +'/address';
	mac = os.popen( cmd).read().strip();
	
	cmd = 'ip -4 addr show '+ dev +' | grep -oP \'(?<=inet\s)\d+(\.\d+){3}\'';
	ip = os.popen( cmd).read().strip();
	
	res = {
		'ip'	:	ip,
		'dev'	:	dev,
		'mac'	:	mac
	};

	return json.dumps( res), 201;	

#------------------------#

@app.route('/api/'+ API_VER +'/system/gwid', methods=['GET'])
def gwid():
	netRes = net_get();
	netInfo = json.loads( netRes[0]);

	conf = {
		'gateway_conf': { 'gateway_ID' : netInfo['mac'].replace(':', '').upper()}
	};
	conf_set( conf);

	return json.dumps( conf['gateway_conf']['gateway_ID']), 201;

#------------------------#

@app.route('/api/'+ API_VER +'/system/wifi/devices', methods=['GET'])
def wifi_devices():
	cmd = 'iw dev | awk \'$1=="Interface"{print $2}\'';
	res = os.popen( cmd).read().strip().split( os.linesep);
	return json.dumps( res), 201;

#------------------------#

@app.route('/api/'+ API_VER +'/system/wifi', methods=['GET'])
def wifi_get():
	
	#cmd = 'ifconfig '+ WIFI_DEVICE +' | grep -oP \'(?<=inet\s)\d+(\.\d+){3}\'';
	cmd = 'ip -4 addr show '+ WIFI_DEVICE +' | grep -oP \'(?<=inet\s)\d+(\.\d+){3}\'';
	ip = os.popen( cmd).read().strip();
	
	#cmd = 'sudo ifconfig | awk \'{print $1}\' | grep "'+ WIFI_DEVICE +':"';
	#cmd = 'nmcli connection show --active | awk \'{print $4}\' | grep "'+ WIFI_DEVICE +'"';
	cmd = 'cat /proc/net/wireless | grep '+ WIFI_DEVICE;
	enabled = len( os.popen( cmd).read().strip()) > 0;	

	#cmd = 'iwconfig '+ WIFI_DEVICE +' | grep SSID | awk \'{match($0,/ESSID:"([^\"]+)"/,a)}END{print a[1]}\'';
	cmd = 'iw '+ WIFI_DEVICE +' info | grep ssid | awk \'{print $2}\'';
	ssid = os.popen( cmd).read().strip();
	
	res = {
		'ip'		:	ip,
		'enabled'	:	enabled,
		'ssid'		:	ssid
	};
	
	return json.dumps( res), 201;	

#------------------------#

@app.route( '/api/'+ API_VER +'/system/wifi', methods=['PUT', 'POST'])
def wifi_set():
	if( request.json is None):
		return "JSON Err!", 201;

	res = [];
	
	if( 'enabled' in request.json):
		if( request.json['enabled'] == '1' or request.json['enabled'] == True):
			status = 'connect';
		else:
			status = 'disconnect';
		cmd = 'nmcli dev '+ status +' "'+ WIFI_DEVICE +'" ';
		res.append( os.popen( cmd).read().strip());
		#iface wlx0013eff1186f inet manual in /etc/network/interfaces
		
	if( 'ssid' in request.json):
		print( os.popen( 'ip link set '+ WIFI_DEVICE +' up').read());
		
		print( os.popen( 'nmcli connection delete id "'+ request.json['ssid'] +'"').read()); #avoid duplication
		print( os.popen( 'nmcli connection up ifname '+ WIFI_DEVICE +'').read());

		cmd  = 'nmcli dev wifi connect "'+ request.json['ssid'] +'"'
		if( len( request.json['password']) >= 8):
			cmd += ' password "'+ request.json['password'] +'"';
		cmd += ' ifname '+ WIFI_DEVICE +' ';
		
		res.append( os.popen( cmd).read());


	return json.dumps( res), 201;

#------------------------#

@app.route('/api/'+ API_VER +'/system/wifi/scanning', methods=['GET'])
@app.route('/api/'+ API_VER +'/system/wifi/scan', methods=['GET'])
def wifi_scan():
	#
	#cmd = 'nmcli -f SSID,SIGNAL,SECURITY dev wifi list ifname '+ WIFI_DEVICE;
	cmd = 'iw '+ WIFI_DEVICE +' scan | awk -f '+ PATH +'/scan.awk'; #| sort -k1,1 -u
	#lines = os.popen( cmd).read().strip().split( os.linesep);
	lines = os.popen( cmd).read().strip();
	lines = lines.split( os.linesep);

	res = []
	for ln in lines:
		wrd = ln.split();
		if( len( wrd) == 3):
			rw = {
				'name'		: wrd[0],
				'signal'	: wrd[1],
				'security'	: wrd[2]
			};
			res.append( rw);

	return json.dumps( res), 201;

#------------------------#

@app.route('/api/'+ API_VER +'/system/wifi/ssid', methods=['POST'])
def wifi_save_ssid():
#	cmd = SCRIPTS +'prepare_wifi_client.sh';
#	print( os.popen( cmd).read());

	print( os.popen( 'ip link set '+ WIFI_DEVICE +' up').read());
	
	print( os.popen( 'nmcli connection delete id "'+ request.json['ssid'] +'"').read()); #avoid duplication
	print( os.popen( 'nmcli connection up ifname '+ WIFI_DEVICE +'').read());

	cmd  = 'nmcli dev wifi connect "'+ request.json['ssid'] +'"'
	if( len( request.json['password']) >= 8):
		cmd += ' password "'+ request.json['password'] +'"';
	cmd += ' ifname '+ WIFI_DEVICE +' ';
	
	res = os.popen( cmd).read();

	return json.dumps( res), 201;

#------------------------#

@app.route( '/api/'+ API_VER +'/system/wifi/ap', methods=['GET'])
def wifi_get_ap():

	cmd = 'egrep "^ssid=" /etc/hostapd/hostapd.conf | awk \'{match($0, /ssid=([^\"]+)/, a)} END{print a[1]}\'';
	ssid = os.popen( cmd).read().strip();

	cmd = 'egrep "^wpa_passphrase=" /etc/hostapd/hostapd.conf | awk \'{match($0, /wpa_passphrase=([^\"]+)/, a)} END{print a[1]}\'';
	password = os.popen( cmd).read().strip();
	
	res = {
		'SSID'		: ssid,
		'password'	: password
	};
	
	return json.dumps( res), 201;

#https://gist.github.com/narate/d3f001c97e1c981a59f94cd76f041140

#------------------------#


@app.route( '/api/'+ API_VER +'/system/wifi/ap', methods=['PUT', 'POST'])
def wifi_set_ap_api():
	if( request.json is None):
		return "", 201;

	return json.dumps( wifi_set_ap( request.json)), 201;

#.............#

def wifi_set_ap( req):

	res = [];
	if( 'SSID' in req):
		#Replacing hot-spot ssid in /etc/hostapd/hostapd.conf
		os.popen( 'sed -i \'s/^ssid.*/ssid='+ req['SSID'] +'/g\' /etc/hostapd/hostapd.conf').read();
		#indicate that a custom ssid has been defined by the user
		os.popen( 'echo '+ req['SSID'] +' | tee /etc/hostapd/custom_ssid.txt > /dev/null').read();
		res.append( 'SSID saved.');

	if( 'password' in req):
		#Setting wpa_passphrase in /etc/hostapd/hostapd.conf
		os.popen( 'sed -i \'s/^wpa_passphrase.*/wpa_passphrase='+ req['password'] +'/g\' /etc/hostapd/hostapd.conf').read();
		res.append( "Password saved.");
	
	if( 'interface' in req):
		#Setting wpa_passphrase in /etc/hostapd/hostapd.conf
		os.popen( 'sed -i \'s/^interface.*/interface='+ req['interface'] +'/g\' /etc/hostapd/hostapd.conf').read();
		res.append( "Interface saved.");

	#print( res);

	return res;

#------------------------#

@app.route( '/api/'+ API_VER +'/system/<status>', methods=['PUT'])
def system_shutdown( status):

	if( status == 'reboot'):
		cmd = 'shutdown -r now';

	else:
		if( status == 'shutdown'):
			cmd = 'shutdown -h now';
		else:
			return 1, 201
	
	res = cmd;
	res = os.popen( cmd).read();
	return json.dumps( res), 201;

#------------------------#

@app.route( '/api/'+ API_VER +'/system/logs', methods=['GET'])
def get_logs_api():
	return json.dumps( get_logs()), 201;

#------------#

@app.route( '/api/'+ API_VER +'/system/logs500', methods=['GET'])
def get_logs_500_api():
	return json.dumps( get_logs( 500)), 201;

#------------#

@app.route( '/api/'+ API_VER +'/system/logs50', methods=['GET'])
def get_logs_50_api():
	return json.dumps( get_logs( 50)), 201;

#------------#

def get_logs( n = 0):
	if( n > 0):
		cmd = 'tail -n '+ str( n ) +' '+ LOGS_PATH +'/post-processing.log';
	else:
		cmd = 'cat '+ LOGS_PATH +'/post-processing.log';

	return os.popen( cmd).read();

#------------------------#

@app.route( '/api/'+ API_VER +'/remote.it', methods=['GET'])
def remoteIT():

	doneF = PATH + '/remote.it/done.txt';
	if( os.path.isfile( doneF) == False):
		return json.dumps( False), 201;

	try:
		with open( doneF) as f:
			regId = f.read();
		res = {
			'time'	: time.ctime( os.path.getmtime( doneF)),
			'id'	: regId
		};

	except OSError:
		res = 0

	
	return json.dumps( res), 201;

#------------------------#

if __name__ == "__main__":
	debugMode	= os.environ['DEBUG_MODE'] == '1';
	apiAddr		= os.environ['WAZIGATE_SYSTEM_ADDR'];
	
	addr = apiAddr.split(':');
	
	apiHost = "0.0.0.0";
	apiPort = 5000;

	if( len( addr) > 0 in addr and len( addr[0]) > 0):
		apiHost = addr[0];

	if( len( addr) > 1 in addr and len( addr[1]) > 0):
		apiPort = int( addr[1]);

	app.run( host = apiHost, debug = debugMode, port = apiPort);


package api

import (
	// "fmt"
	"net"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	routing "github.com/julienschmidt/httprouter"
	
	// In future we will use something like this lib to handle wifi stuff
	// wifi "github.com/mark2b/wpa-connect"
)

/*-------------------------*/

func GetNetInfo( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	
	dev, _ := exeCmd( "ip route show default | head -n 1 | awk '/default/ {print $5}'")
	mac, _ := exeCmd( "cat /sys/class/net/"+ dev +"/address")
	
	
	cmd := "ip -4 addr show "+ dev +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}'";
	ip, _ := exeCmd( cmd)
	
	/*---------------*/

	out := map[string]interface{}{
		"ip"	:	ip,
		"dev"	:	dev,
		"mac"	:	mac,
	}

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}

	resp.Write( []byte( outJson))
}

/*-------------------------*/

func GetGWID( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	interfs, err := net.Interfaces()
	if err != nil {
		log.Printf( "[Err   ] %s", err.Error())

		if( DEBUG_MODE){ 
			http.Error( resp, "[ Error ]: "+ err.Error(), http.StatusInternalServerError)
		}
	}

	localID := ""

	for _, interf := range interfs {
		addr := interf.HardwareAddr.String()
		if addr != "" {
			localID = strings.ReplaceAll( addr, ":", "")
			break
		}
	}
	resp.Write( []byte( localID))
}

/*-------------------------*/

func GetNetWiFi( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	
	/*iwifi, err := net.InterfaceByName( WIFI_DEVICE)
	if err != nil {
		log.Printf( "[Err   ] %s", err.Error())
		if( DEBUG_MODE){ 
			http.Error( resp, "[ Error ]: "+ err.Error(), http.StatusInternalServerError)
		}

	}

	addrs, err := iwifi.Addrs();
	if err != nil {
		log.Printf( "[Err   ] %s", err.Error())
		if( DEBUG_MODE){ 
			http.Error( resp, "[ Error ]: "+ err.Error(), http.StatusInternalServerError)
		}
	}

	ip := ""
	if len( addrs) > 0 {
		ip = addrs[0].(*net.IPNet).IP.String()
	} /**/

	/*-----*/

	cmd := "ip -4 addr show "+ WIFI_DEVICE +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}'";
	ip, _ := exeCmd( cmd);

	/*-----*/
	
	cmd = "ip link show up "+ WIFI_DEVICE;
	outc, _ := exeCmd( cmd)
	enabled := outc != "";
	
	/*-----*/

	cmd = "iw "+ WIFI_DEVICE +" info | grep ssid | awk '{print $2\" \"$3\" \"$4\" \"$5\" \"$6}'";
	outc, _ = exeCmd( cmd)
	ssid := outc

	/*-----*/

	cmd = "systemctl is-active --quiet hostapd && echo 1"
	outc, _ = execOnHost( cmd)
	ap_mode := outc == "1"

	/*---------------*/

	out := map[string]interface{}{
		"ip"		:	ip,
		"enabled"	:	enabled,
		"ssid"		:	ssid,
		"ap_mode"	:	ap_mode,
	}

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}

	resp.Write( []byte( outJson))
}

/*-------------------------*/

func SetNetWiFi( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	
	if err := req.ParseForm(); err != nil {
		log.Printf( "[Err   ] %s", err.Error())
		if( DEBUG_MODE){ 
			http.Error( resp, "[ Error ]: "+ err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var reqJson map[string]interface{}
	decoder := json.NewDecoder( req.Body)
    err := decoder.Decode( &reqJson)
    if err != nil {
        log.Printf( "[Err   ] %s", err.Error())
    }

	// log.Println( reqJson)

	if enabled, exist := reqJson["enabled"]; exist{
		if enabled == true || enabled == "1"{
			exeCmd( "ip link set "+ WIFI_DEVICE +" up")
		}else{
			exeCmd( "ip link set "+ WIFI_DEVICE +" down")
		}
	}

	if ssid, exist := reqJson["ssid"]; exist{
		exeCmd( "ip link set "+ WIFI_DEVICE +" up")
	
		cmd := "sudo cp /etc/wpa_supplicant/wpa_supplicant.conf.orig /etc/wpa_supplicant/wpa_supplicant.conf;"
		// exeCmd( cmd)
		execOnHost( cmd)
		
		cmd = ""
		if str, ok := ssid.(string); ok{
			cmd += "sudo wpa_passphrase \""+ str +"\"";
		}

		if password, exist := reqJson["password"]; exist{
			if str, ok := password.(string); ok{
				cmd += " \""+ str +"\""
			}
		}

		cmd += " >> /etc/wpa_supplicant/wpa_supplicant.conf; ";
		// exeCmd( cmd)
		execOnHost( cmd)
		
		// save the setting and switch to the WiFi Client

		oledWrite( "\nConnecting to\n   WiFi...")
		stdout, _ := execOnHost( "sudo bash start_wifi.sh")
		log.Printf( "[Info   ] %s", stdout)
		oledWrite( "") // Clean the OLED msg

		CheckWlanConn() // Check if the WiFi connection was successfull otherwise revert to AP mode
	}

	out := "WiFi set successfully";

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}

	resp.Write( []byte( outJson))

	// resp.Write( []byte( "WiFi configs set successfully"))
}

/*-------------------------*/

func apMode( withLogs bool) bool {

	apAtive, _ := execOnHostWithLogs( "systemctl is-active --quiet hostapd && echo 1", withLogs)
	return apAtive == "1"
}

/*-------------------------*/

func CheckWlanConn() bool{

	if apMode( true){
		ActivateAPMode()
		return false // The AP Mode is active
	}

	//In WLAN Mode:
	time.Sleep( 3 * time.Second)
	for i := 0; i < 4; i++ {

		oledWrite( "\nChecking WiFi." + strings.Repeat( ".", i));
		
		wifiRes, _ := execOnHost( "iwgetid")
		if wifiRes != ""{
			oledWrite( ""); // Clean the OLED
			return true
		}

		time.Sleep( 3 * time.Second)
		oledWrite( "");
		time.Sleep( 2 * time.Second)
	}
	
	//Could no conenct, need to revert to AP setting
	
	if( DEBUG_MODE){
		log.Printf( "[Info  ] Could not connect!\nReverting the settings...")
	}
	oledWrite( "Cannot Connect\n\nReverting to \n  Access point...")
	time.Sleep( 2 * time.Second)
	
	ActivateAPMode()

	return true
}

/*-------------------------*/

func ActivateAPMode() {

	oledWrite( "\nActivating\n Access point mode...");

	stdout, _ := execOnHost( "sudo bash start_hotspot.sh")
	if( DEBUG_MODE){
		log.Printf( "[Info  ] %s", stdout)
	}

	oledWrite( ""); // Clean the OLED

	time.Sleep( 1 * time.Second)
}

/*-------------------------*/

func SetNetAPMode( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	ActivateAPMode()

	out := "Access Point mode Activated.";

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}

	resp.Write( []byte( outJson))	
	// resp.Write( []byte( "OK"))
}

/*-------------------------*/

func NetWiFiScan( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	
	cmd := "iw "+ WIFI_DEVICE +" scan | awk -f scan.awk"
	out, _ := exeCmd( cmd)
	lines := strings.Split( out, "\n")

	resp.Header().Set("Content-Type", "application/json")
	resp.Write([]byte{'['})

	firstItemServed := false
	for _, line := range lines {
		wrd := strings.Split( string( line), "\t")
		if len( wrd) == 3 && wrd[0] != "" {

			if firstItemServed {
				resp.Write([]byte{','})	
			}
			
			out := map[string]interface{}{
				"name"		:	wrd[0],
				"signal"	:	wrd[1],
				"security"	:	wrd[2],
			}

			outJson, err := json.Marshal( out)
			if( err != nil) {
				log.Printf( "[Err   ] %s", err.Error())
			}

			if( DEBUG_MODE){
				log.Printf( "[Info  ] WiFi Scan: %v", out)
			}
		
			resp.Write( []byte( outJson))
			
			firstItemServed = true
		}
	}

	resp.Write([]byte{']'})	
}

/*-------------------------*/

func GetNetAP( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	var cmd string

	cmd = "egrep \"^ssid=\" /etc/hostapd/hostapd.conf | awk '{match($0, /ssid=([^\"]+)/, a)} END{print a[1]}'"
	ssid, _ := execOnHost( cmd)

	cmd = "egrep \"^wpa_passphrase=\" /etc/hostapd/hostapd.conf | awk '{match($0, /wpa_passphrase=([^\"]+)/, a)} END{print a[1]}'"
	password, _ := execOnHost( cmd)
	
	cmd = "iw dev | awk '$1==\"Interface\"{print $2}' | grep \""+ WIFI_DEVICE +"\""
	deviceRes, _ := exeCmd( cmd)
	
	cmd = "ip -4 addr show "+ WIFI_DEVICE +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}'"
	ip, _ := exeCmd( cmd)

	out := map[string]interface{}{
		"available"	:	deviceRes != "",
		"device"	:	WIFI_DEVICE,
		"SSID"		:	ssid,
		"password"	:	password,
		"ip"		:	ip,
	}

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}

	resp.Write( []byte( outJson))

}

/*-------------------------*/

func SetNetAP( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	if err := req.ParseForm(); err != nil {
		log.Printf( "[Err   ] %s", err.Error())
		if( DEBUG_MODE){ 
			http.Error( resp, "[ Error ]: "+ err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var reqJson map[string]interface{}
	decoder := json.NewDecoder( req.Body)
    err := decoder.Decode( &reqJson)
    if err != nil {
        log.Printf( "[Err   ] %s", err.Error())
    }

	var cmd string

	out := ""
	
	if ssid, exist := reqJson["SSID"]; exist{
		if str, ok := ssid.(string); ok{
			cmd = "sed -i 's/^ssid.*/ssid="+ str +"/g' /etc/hostapd/hostapd.conf"
			execOnHost( cmd)

			// cmd = "echo "+ str +" | tee /etc/hostapd/custom_ssid.txt > /dev/null"
			// exeCmd( cmd)

			out += "SSID ";
		}
	}

	if password, exist := reqJson["password"]; exist{
		if str, ok := password.(string); ok{
			cmd = "sed -i 's/^wpa_passphrase.*/wpa_passphrase="+ str +"/g' /etc/hostapd/hostapd.conf"
			execOnHost( cmd)

			out += "and Password ";
		}
	}


	out += "saved.";

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}

	resp.Write( []byte( outJson))	

}

/*-------------------------*/

func CloudAccessible( withLogs bool) bool{

	cmd := "timeout 3 curl -Is https://waziup.io | head -n 1 | awk '{print $2}'"
	rCode, _ := exeCmdWithLogs( cmd, withLogs)

	return rCode == "200"
}

/*-------------------------*/


func InternetAccessible( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	if CloudAccessible( true) {
		
		resp.Write( []byte( "1"))

	}else{

		resp.Write( []byte( "0")) 
	}

}

/*-------------------------*/

func GetAllIPs() (string, string, string, string) {

	cmd := "iw "+ WIFI_DEVICE +" info | grep ssid | awk '{print $2\" \"$3\" \"$4\" \"$5\" \"$6}'";
	ssid, _ := exeCmdWithLogs( cmd, false)
	
	cmd = "status=$(ip addr show "+ WIFI_DEVICE +" | grep \"state UP\"); if [ \"$status\" == \"\" ]; then echo \"\"; else echo $(ip -4 addr show "+ WIFI_DEVICE +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}');  fi;"
	wip, _ := exeCmdWithLogs( cmd, false)
	aip := wip

	cmd = "status=$(ip addr show "+ ETH_DEVICE +" | grep \"state UP\"); if [ \"$status\" == \"\" ]; then echo \"NO Ethernet\"; else echo $(ip -4 addr show "+ ETH_DEVICE +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}');  fi;"
	eip, _ := exeCmdWithLogs( cmd, false)

	if apMode( false){
		wip = ""
	}else{
		aip = ""
	}

	return eip, wip, aip, ssid
}

/*-------------------------*/
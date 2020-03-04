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
	
	dev := exeCmd( "ip route show default | head -n 1 | awk '/default/ {print $5}'", resp)
	mac := exeCmd( "cat /sys/class/net/"+ dev +"/address", resp)
	
	
	cmd := "ip -4 addr show "+ dev +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}'";
	ip := exeCmd( cmd, resp)
	
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
	ip := exeCmd( cmd, resp);

	/*-----*/
	
	cmd = "ip link show up "+ WIFI_DEVICE;
	enabled := exeCmd( cmd, resp) != "";
	
	/*-----*/

	cmd = "iw "+ WIFI_DEVICE +" info | grep ssid | awk '{print $2\" \"$3\" \"$4\" \"$5\" \"$6}'";
	ssid := exeCmd( cmd, resp);

	/*-----*/

	cmd = "systemctl is-active --quiet hostapd && echo 1"
	ap_mode := execOnHost( cmd, resp) == "1"

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
			exeCmd( "ip link set "+ WIFI_DEVICE +" up", resp)
		}else{
			exeCmd( "ip link set "+ WIFI_DEVICE +" down", resp)
		}
	}

	if ssid, exist := reqJson["ssid"]; exist{
		exeCmd( "ip link set "+ WIFI_DEVICE +" up", resp)
	
		cmd := "sudo cp /etc/wpa_supplicant/wpa_supplicant.conf.orig /etc/wpa_supplicant/wpa_supplicant.conf;"
		// exeCmd( cmd, resp)
		execOnHost( cmd, resp)
		
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
		// exeCmd( cmd, resp)
		execOnHost( cmd, resp)
		
		// save the setting and switch to the WiFi Client

		oledWrite( "\nConnecting to\n   WiFi...")
		stdout := execOnHost( "sudo bash start_wifi.sh", resp)
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

	apAtive := execOnHostWithLogs( "systemctl is-active --quiet hostapd && echo 1", withLogs, nil)
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
		
		wifiRes := execOnHost( "iwgetid", nil)
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

	stdout := execOnHost( "sudo bash start_hotspot.sh", nil)
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
	lines := strings.Split( exeCmd( cmd, resp), "\n")

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
	ssid := execOnHost( cmd, resp)

	cmd = "egrep \"^wpa_passphrase=\" /etc/hostapd/hostapd.conf | awk '{match($0, /wpa_passphrase=([^\"]+)/, a)} END{print a[1]}'"
	password := execOnHost( cmd, resp)
	
	cmd = "iw dev | awk '$1==\"Interface\"{print $2}' | grep \""+ WIFI_DEVICE +"\""
	deviceRes := exeCmd( cmd, resp)
	
	cmd = "ip -4 addr show "+ WIFI_DEVICE +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}'"
	ip := exeCmd( cmd, resp)

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
			execOnHost( cmd, resp)

			// cmd = "echo "+ str +" | tee /etc/hostapd/custom_ssid.txt > /dev/null"
			// exeCmd( cmd, resp)

			out += "SSID ";
		}
	}

	if password, exist := reqJson["password"]; exist{
		if str, ok := password.(string); ok{
			cmd = "sed -i 's/^wpa_passphrase.*/wpa_passphrase="+ str +"/g' /etc/hostapd/hostapd.conf"
			execOnHost( cmd, resp)

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
	rCode := exeCmdWithLogs( cmd, withLogs, nil)

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
	ssid := exeCmdWithLogs( cmd, false, nil)
	
	cmd = "status=$(ip addr show "+ WIFI_DEVICE +" | grep \"state UP\"); if [ \"$status\" == \"\" ]; then echo \"\"; else echo $(ip -4 addr show "+ WIFI_DEVICE +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}');  fi;"
	wip := exeCmdWithLogs( cmd, false, nil)
	aip := wip

	cmd = "status=$(ip addr show "+ ETH_DEVICE +" | grep \"state UP\"); if [ \"$status\" == \"\" ]; then echo \"NO Ethernet\"; else echo $(ip -4 addr show "+ ETH_DEVICE +" | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}');  fi;"
	eip := exeCmdWithLogs( cmd, false, nil)

	if apMode( false){
		wip = ""
	}else{
		aip = ""
	}

	return eip, wip, aip, ssid
}

/*-------------------------*/
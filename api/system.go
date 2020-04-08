package api

import (
	// "fmt"
	"encoding/json"
	"log"
	"time"
	"strings"
	"regexp"
	// "strconv"
	"io"
	"context"
	"net"
	"net/http"
	
	"os"
	"os/exec"
	"path/filepath"
	"io/ioutil"

	routing "github.com/julienschmidt/httprouter"
)

/*-------------------------*/


type Configuration struct {
	Setup_wizard 			bool	`json:"setup_wizard"`
	WiFi_AP_auto			bool	`json:"wifi_ap_auto"`	 		// Check if WiFi is not connected, wait for n seconds then switch to AP mode and keep searching
	WiFi_AP_no_Internet		bool	`json:"wifi_ap_no_internet"`	// Check if there is no internet, switch to AP mode or try to find another WiFi
	WiFi_timeout			int		`json:"wifi_timeout"`			// How many seconds wait before doing an action on the WiFi
	Fan_trigger_temp		float64	`json:"fan_trigger_temp"`		// At which temperature the Fan should start
	OLED_halt_timeout		int		`json:"oled_halt_timeout"`		// After what time the OLED goes off
}

/*----------------*/

func loadConfigs() Configuration {

	filename := GetRootPath() +"/conf.json"
	bytes, err := ioutil.ReadFile( filename)
	if err != nil {
		log.Printf( "[Err   ] %s", err.Error())
		return Configuration{
			false,
			true,
			true,
			60,		
			62.1,	// in CC
			1 * 60, // 5 minutes
		}
	}

	var c Configuration
	err = json.Unmarshal( bytes, &c)
	if err != nil {
		log.Printf( "[Err   ] %s", err.Error())
		return Configuration{}
	}
	return c
}

/*-------------------------*/

func exeCmdWithLogs( cmd string, withLogs bool) ( string, error) {

	if( withLogs && DEBUG_MODE){
		log.Printf( "[Info  ] executing [ %s ] ", cmd)
	}

	exe := exec.Command( "sh", "-c", cmd)
    stdout, err := exe.Output()

    if( err != nil) {
		if( withLogs){
			log.Printf( "[Err   ] executing [ %s ] command. \n\tError: [ %s ]", cmd, err.Error())
		}
        return "", err
	}
	return strings.Trim( string( stdout), " \n\t\r"), nil
}

/*-------------------------*/

func exeCmd( cmd string) ( string, error) {
	return exeCmdWithLogs( cmd, true)
}

/*-------------------------*/

func execOnHostWithLogs( cmd string, withLogs bool) ( string, error) {

	if( withLogs && DEBUG_MODE){
		log.Printf( "[Exec  ]: Host Command [ %s ]", cmd)
	}

	socketAddr := os.Getenv( "WAZIGATE_HOST_ADDR")
	if socketAddr == "" {
		socketAddr = "/var/run/wazigate-host.sock" // Default address for the Host
	}

	response, err := SocketReqest(socketAddr, "cmd", "POST", "application/json", strings.NewReader(cmd), withLogs)

	if err != nil {
		if response != nil && response.Body != nil{
			response.Body.Close()
		}
		if( withLogs){
			log.Printf( "[Err   ]: %s ", err.Error())
		}

		oledWrite( "\n\n  HOST ERROR!")

		return "", err
	}

	resBody, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log.Printf( "[Err   ]: %s ", err.Error())
		return "", err
	}
	return string( resBody), nil
}

/*-------------------------*/

func execOnHost( cmd string) (string, error) {
	return execOnHostWithLogs( cmd, true)
}

/*-------------------------*/

func GetRootPath() string {
	dir, err := filepath.Abs( filepath.Dir( os.Args[0]))
    if err != nil {
		log.Fatal( err)
    }
    return dir
}

/*-------------------------*/

func GetSystemConf( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	// bytes, err := json.MarshalIndent( Config, "", "  ")
	bytes, err := json.Marshal( Config)

	if( err != nil){
		log.Printf( "[Err   ] %s", err.Error())

		errorDesc := ""
		if( DEBUG_MODE){ errorDesc = err.Error()}
		if( resp != nil){
			http.Error( resp, "[Err   ]: "+ errorDesc, http.StatusInternalServerError)
		}
	}
	resp.Write( bytes)
}

/*-------------------------*/

func SetSystemConf( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	decoder := json.NewDecoder( req.Body)

	if err := decoder.Decode( &Config); err != nil {

		log.Printf( "[Err   ] %s", err.Error())
		if( DEBUG_MODE){ 
			http.Error( resp, "[ Error ]: "+ err.Error(), http.StatusBadRequest)
		}		
		return
	}

	saveConfig( Config)
	resp.Write( []byte( "Saved"))
}

/*-------------------------*/

func SystemShutdown( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	systemShutdown( "shutdown")
}

/*-------------------------*/

func SystemReboot( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	systemShutdown( "reboot")
}

/*-------------------------*/

func systemShutdown( status string) {

	cmd := "sudo docker stop $(sudo docker ps -a -q); "
	if status == "reboot" {
		oledWrite( "\n  Rebooting...");
		cmd += "sudo shutdown -r now"
	}

	if status == "shutdown" {
		oledWrite( "\nShutting down...");
		cmd += "sudo shutdown -h now"
	}

	time.Sleep( 2 * time.Second)
	
	oledWrite( ""); // Clean the OLED

	time.Sleep( 1 * time.Second)

	log.Printf( "[Info  ] System %s", status)

	oledHalt()

	stdout, _ := execOnHost( cmd)
	log.Printf( "[Info  ] %s", stdout)
}

/*-------------------------*/

func systemQuickShutdown() {

	cmd := "sudo docker stop $(sudo docker ps -a -q); sudo shutdown -h now"

	stdout, _ := execOnHost( cmd)
	log.Printf( "[Info  ] %s", stdout)
}

/*-------------------------*/

func saveConfig( c Configuration) {
	
	filename := GetRootPath() +"/conf.json"

	bytes, err := json.MarshalIndent( c, "", "  ")
	if err != nil {
		log.Printf( "[Err   ] %s", err.Error())
		return
	}

	err = ioutil.WriteFile( filename, bytes, 0644)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}		
}

/*-------------------------*/


func SystemUpdate( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	oledWrite( "\nUpdating...");

	cmd := "sudo bash update.sh | sudo tee update.logs &"; // Run it and unlock the thing

	stdout, _ := execOnHost( cmd);
	log.Printf( "[Info   ] %s", stdout)

	oledWrite( "\nDONE.");

	time.Sleep( 1 * time.Second)

	oledWrite( ""); // Clean the OLED

	out := "Update Done."

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}/**/

	resp.Write( []byte( outJson))
}

/*-------------------------*/

func SystemUpdateStatus( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	cmd := "[ -f update.logs ] && cat update.logs";
	stdout, err := execOnHost( cmd);
	if( err != nil) {
		stdout = ""
		log.Printf( "[Err   ] %s", err.Error())
	}

	outJson, err := json.Marshal( stdout)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}
	resp.Write( []byte( outJson))
}

/*-------------------------*/

func GetGWBootstatus( withLogs bool) ( bool, string){

	cmd := "curl -s --unix-socket /var/run/docker.sock http://localhost/containers/json?all=true"
	outJsonStr, _ := execOnHostWithLogs( cmd, withLogs)

	var resJson []map[string]interface{}

	json.Unmarshal( []byte( outJsonStr), &resJson)

	allOk := true
	out	:= ""

	for _, obj := range resJson {

		// Finding the wazigate containers...
		re := regexp.MustCompile(`/wazigate-(.*)`)
		reFnd := re.FindSubmatch([]byte( obj["Names"].([]interface{})[0].(string)))

		if len( reFnd) < 1 {
			continue
		}
		cName := string( reFnd[1])

		cState := strings.ToUpper( obj["State"].(string))

		if cState != "RUNNING"{
			allOk = false
		}
		// cState = cState[0:3]

		neededSpaces := 16 - len( cName ) - 2 - len( cState)
		out += cName +": "+ strings.Repeat( " ", neededSpaces)+ cState +"\n"
	}

	return allOk, out
}

/*-------------------------*/

func FirmwareVersion( resp http.ResponseWriter, req *http.Request, params routing.Params) {

	out := os.Getenv( "WAZIUP_VERSION")

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}

	resp.Write( []byte( outJson))	

}

/*-------------------------*/

// SocketReqest makes a request to a unix socket
func SocketReqest(socketAddr string, url string, method string, contentType string, body io.Reader, withLogs bool) (*http.Response, error) {

	if( withLogs && DEBUG_MODE){
		log.Printf("[SOCK ] `%s` %s \"%s\" '%v'", socketAddr, method, url, body)
	}
	
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketAddr)
			},
			MaxIdleConns:       50,
			IdleConnTimeout:    3 * 60 * time.Second,
		},
	}

	req, err := http.NewRequest( method, "http://localhost/"+url, body)
	
	if err != nil {
		log.Printf("[Socket   ]: %s ", err.Error())
		return nil, err
	}
	
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	
	response, err := httpc.Do( req)
	// defer response.Body.Close()
	
	if err != nil {
		log.Printf("[Socket]: %s ", err.Error())
		return nil, err
	}

	return response, nil
}

/*-------------------------*/

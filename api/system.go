package api

import (
	// "fmt"
	"encoding/json"
	"log"
	"time"
	"strings"
	"regexp"
	// "strconv"
	"net/http"
	
	"os"
	"os/exec"
	"path/filepath"
	"io/ioutil"

	routing "github.com/julienschmidt/httprouter"
)

/*-------------------------*/

type Configuration struct {
	Setup_wizard    bool `json:"setup_wizard"`
}

/*-------------------------*/

func exeCmdWithLogs( cmd string, withLogs bool, resp http.ResponseWriter) string {

	if( withLogs && DEBUG_MODE){
		log.Printf( "[Info  ] executing [ %s ] ", cmd)
	}

	exe := exec.Command( "sh", "-c", cmd)
    stdout, err := exe.Output()

    if( err != nil) {
		if( withLogs){
			log.Printf( "[Err   ] executing [ %s ] command. \n\tError: [ %s ]", cmd, err.Error())

			errorDesc := ""
			if( DEBUG_MODE){ errorDesc = err.Error()}

			if( resp != nil){
				http.Error( resp, "[Err   ]: "+ errorDesc, http.StatusInternalServerError)
			}
		}
        return ""
	}
	return strings.Trim( string( stdout), " \n\t\r")
}

/*-------------------------*/

func exeCmd( cmd string, resp http.ResponseWriter) string {
	return exeCmdWithLogs( cmd, true, resp)
}

/*-------------------------*/

func execOnHostWithLogs( cmd string, withLogs bool, resp http.ResponseWriter) string {

	if( withLogs && DEBUG_MODE){
		log.Printf( "[Exec  ]: Host Command [ %s ]", cmd)
	}
	
	addr := os.Getenv( "WAZIGATE_HOST_ADDR")
	if addr == "" {
		addr = ":5200" // Default address for the Host
	}
	
	resCall, err := http.Post( "http://"+ addr +"/cmd", "text/html", strings.NewReader( cmd))
	if( err != nil) {
		
		if( withLogs){
			log.Printf( "[Err   ]: %s ", err.Error())

			errorDesc := ""
			if( DEBUG_MODE){ errorDesc = err.Error()}
			if( resp != nil){
				http.Error( resp, "[Err   ]: "+ errorDesc, http.StatusInternalServerError)
			}
		}

		oledWrite( "\n\n  HOST ERROR!")

		return ""
	}

	body, err := ioutil.ReadAll( resCall.Body)
	if( err != nil) {
		if( withLogs){
			log.Printf( "[Err   ]: %s ", err.Error())
			
			errorDesc := ""
			if( DEBUG_MODE){ errorDesc = err.Error()}
			if( resp != nil){
				http.Error( resp, "[Err   ]: "+ errorDesc, http.StatusInternalServerError)
			}
		}
		return ""
	}	

	return string( body)

}

/*-------------------------*/

func execOnHost( cmd string, resp http.ResponseWriter) string {
	return execOnHostWithLogs( cmd, true, resp)
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

	stdout := execOnHost( cmd, nil)
	log.Printf( "[Info  ] %s", stdout)
}


/*-------------------------*/

func loadConfigs() Configuration {

	filename := GetRootPath() +"/conf.json"
	bytes, err := ioutil.ReadFile( filename)
	if err != nil {
		log.Printf( "[Err   ] %s", err.Error())
		return Configuration{}
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

	stdout := execOnHost( cmd, resp);
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
	stdout := execOnHost( cmd, resp);

	outJson, err := json.Marshal( stdout)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}
	resp.Write( []byte( outJson))
}

/*-------------------------*/

func GetGWBootstatus( withLogs bool) ( bool, string){

	cmd := "curl -s --unix-socket /var/run/docker.sock http://localhost/containers/json?all=true"
	outJsonStr := execOnHostWithLogs( cmd, withLogs, nil)

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

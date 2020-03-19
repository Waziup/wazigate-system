package api

import (
	// "fmt"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	// "time"
	"io/ioutil"

	routing "github.com/julienschmidt/httprouter"
)

/*-------------------------*/

func DockerStatus( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	cmd := "curl --unix-socket /var/run/docker.sock http://localhost/containers/json?all=true";
	outJson := execOnHost( cmd, resp);
	resp.Write( []byte( outJson))

	//Ref: https://docs.docker.com/engine/api/v1.26/	
}

/*-------------------------*/

func DockerStatusById( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	
	//TODO: it returns only the running containers ! Need to be fixed.
	cId	:=	params.ByName( "cId")
	
	qry := url.QueryEscape( "{\"id\":[\""+ cId +"\"]}");
	cmd := "curl --unix-socket /var/run/docker.sock http://localhost/containers/json?filters="+ qry;
	outJson := execOnHost( cmd, resp);
	
	resp.Write( []byte( outJson))

	//Ref: https://docs.docker.com/engine/api/v1.26/	
}

/*-------------------------*/

func DockerAction( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	
	cId		:=	params.ByName( "cId")
	action	:=	params.ByName( "action")

	cmd := "curl --no-buffer -XPOST --unix-socket /var/run/docker.sock http://localhost/containers/"+ cId +"/"+ action;
	out := execOnHost( cmd, resp);
	
	out += " [ "+ action +" ] done.";

	outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}

	resp.Write( []byte( outJson))
}

/*-------------------------*/

func DockerLogs( resp http.ResponseWriter, req *http.Request, params routing.Params) {
	
	cId		:=	params.ByName( "cId")
	tail	:=	params.ByName( "tail")

	cmd := "sudo docker logs -t "+ cId;
	if tail != ""{
		cmd = "sudo docker logs -t --tail="+ tail +" "+ cId;
	}
	out := execOnHostWithLogs( cmd, false, resp);

	/*outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}/**/

	resp.Write( []byte( out))
}

/*-------------------------*/

var DockerInstallAppStatus = "";

func DockerInstallApp( resp http.ResponseWriter, req *http.Request, params routing.Params){

	// imageName := "waziup/wazi-on-sensors:1.0.0"
	input, err := ioutil.ReadAll( req.Body)
	if( err != nil) {
		log.Printf( "[Err   ] installing app [%v] error: %s ", input, err.Error())
		http.Error( resp, err.Error(), http.StatusBadRequest)
		return
	}
	imageName := string( input);
	
	DockerInstallAppStatus = "Installing initialized\n"
	
	//<!-- Get the App information

	appsDir := "../apps/"  //We may use env vars in future, this path is relative to wazigate-host

	sp1 := strings.Split( imageName, ":")
	
	version := ""
	if( len( sp1) == 2){
		version = sp1[1]; //Image version
	}
	
	sp2 := strings.Split( sp1[0], "/")

	repoName:= sp2[0];
	appName	:= repoName + "_app"; // some random default name in case of error
	if( len( sp2) > 1){
		appName = sp2[1]
	}

	appFullPath := appsDir + repoName +"/"+ appName;

	//-->

	DockerInstallAppStatus += "Downloading [ "+ appName +" : "+ version +" ] \n";

	cmd := "docker pull "+ imageName;
	out := execOnHostWithLogs( cmd, true, resp);

	// Status: Downloaded newer image for waziup/wazi-on-sensors:1.0.0
	DockerInstallAppStatus += out;

	if( strings.Contains( out, "Error")){
		resp.WriteHeader(400)
		resp.Write( []byte( "Download Failed!"))
		return
	}

	cmd = "docker create "+ imageName;
	cId := execOnHostWithLogs( cmd, true, resp);

	DockerInstallAppStatus += "Termporary container created\n";

	cmd = "mkdir -p "+ appsDir + repoName;
	cmd += "; mkdir -p "+ appFullPath;
	cmd += "; docker cp "+ cId +":/index.zip "+ appFullPath;
	out = execOnHostWithLogs( cmd, true, resp);

	DockerInstallAppStatus += out;

	// Error: No such container:path....

	if( strings.Contains( out, "Error")){
		resp.WriteHeader(400)
		resp.Write( []byte( "`index.zip` file extraction failed!"))
		return
	}

	cmd = "docker rm "+ cId;
	out = execOnHostWithLogs( cmd, true, resp);

	cmd = "unzip -o "+ appFullPath + "/index.zip -d "+ appFullPath;
	out = execOnHostWithLogs( cmd, true, resp);

	if( strings.Contains( out, "cannot find")){
		resp.WriteHeader(400)
		resp.Write( []byte( "Could not unzip `index.zip`!"))
		return
	}

	/*outJson, err := json.Marshal( out)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}/**/
	
	DockerInstallAppStatus = ""	//Clear the msg buffer
	resp.Write( []byte( out))

}

func DockerInstallAppGetStatus( resp http.ResponseWriter, req *http.Request, params routing.Params){

	resp.Write( []byte( DockerInstallAppStatus))
	// DockerInstallAppStatus = ""
}

/*-------------------------*/

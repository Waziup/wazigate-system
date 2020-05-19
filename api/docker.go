package api

import (
	// "fmt"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	// "strings"
	// "time"

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

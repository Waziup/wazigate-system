import * as React from "react";
import * as ReactDOM from "react-dom";
import * as waziup from "waziup";


import { version, branch } from "./version";
import AppComp from "./components/App";

console.log("This is Wazigate-System, a %s build. %s", branch, version);

// basic UI styles, platform dependant
if (navigator.platform.indexOf("Win") == 0)
	document.body.classList.add("windows");
else if (navigator.platform.indexOf("Mac") == 0)
	document.body.classList.add("mac");
else if (navigator.platform.indexOf("Linux") != -1)
	document.body.classList.add("linux");

//React does not load the corresponding CSS, fix it later
const loader = document.querySelector(".loader");
// if you want to show the loader when React loads data again
const showLoader = () => loader.classList.remove("loader--hide");
const hideLoader = () => loader.classList.add("loader--hide");

if (window.parent && "waziup" in window.parent) {
	(window as any)["wazigate"] = (window.parent as any)["wazigate"];
	console.log("Using wazigate instance from parent window.");
	render();
} else {
	console.log("Connecting to wazigate ...");
	waziup.connect({host: "."}).then(wazigate => {
		(window as any)["wazigate"] = wazigate;
		wazigate.connectMQTT(() => {
			console.log("MQTT Connected.");
		}, (err: Error) => {
			console.error("MQTT Err", err);
		}, {
			reconnectPeriod: 0,
		});
		render();
	});
}

function render() {
	ReactDOM.render(
		<AppComp hideLoader={hideLoader} showLoader={showLoader} />,
		document.getElementById("app")
	);
}





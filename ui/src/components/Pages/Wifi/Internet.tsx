import * as React from "react";
import * as API from "../../../api";
import ErrorComp from "../../Error";

import {
  MDBContainer,
  MDBTabPane,
  MDBTabContent,
  MDBNav,
  MDBNavItem,
  MDBNavLink,
  MDBIcon,
  MDBListGroup,
  MDBAlert,
} from "mdbreact";
import LoadingSpinner from "../../LoadingSpinner";
import WiFiScanItem from "./WiFiScanItem";

declare function Notify(msg: string): any;

export interface Props {
  devices: API.Devices
}
export interface State {
  WiFiScanResults: API.AccessPoint[];
  WiFiInfo: API.Device;
  error: any;
  scanLoading: boolean;
  activeItem: any;
}

class PagesInternet extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      WiFiScanResults: null,
      WiFiInfo: null,
      error: null,
      scanLoading: true,
      activeItem: "wifi",
    };
    console.log("PagesInternet", props);
  }

  /**------------- */
  _isMounted = false;
  componentDidMount() {
    this._isMounted = true;
    this.scan();
    // if( !this._isMounted) return;
  }
  componentWillUnmount() {
    this._isMounted = false;
  }
  /**------------- */

  scan() {
    if (!this._isMounted) return;
    if (this.state.activeItem != "wifi") return;

    this.updateWiFiInfo();

    this.setState({
      scanLoading: true,
    });

    API.getWiFiScan().then(
      (WiFiScanResults) => {
        if (WiFiScanResults.length) {
          let uniqueList = [];
          let tmpArr = Array();

          for (var i = 0; i < WiFiScanResults.length; i++) {
            if (tmpArr.indexOf(WiFiScanResults[i].ssid) == -1) {
              uniqueList.push(WiFiScanResults[i]);
              tmpArr.push(WiFiScanResults[i].ssid);
            }
          }

          this.setState({
            WiFiScanResults: uniqueList,
          });
        }

        this.setState({
          scanLoading: false,
          error: null,
        });

        setTimeout(() => {
          this.scan();
        }, 10000); // 10 seconds
      },
      (error) => {
        this.setState({
          scanLoading: false,
          error: error,
        });

        setTimeout(() => {
          this.scan();
        }, 10000); // 10 seconds
      }
    );
  }

  /**------------- */

  toggle = (tab: any) => () => {
    if (this.state.activeItem !== tab) {
      this.setState({
        activeItem: tab,
      });

      if (tab == "wifi") {
        this.scan();
      }
    }
  };

  /**------------- */

  updateWiFiInfo() {
    API.getWlanDevice().then(
      (WiFiInfo) => {
        // console.log(WiFiInfo);
        this.setState({
          WiFiInfo: WiFiInfo,
          error: null,
        });
      },
      (error) => {
        this.setState({
          WiFiInfo: null,
          error: error,
        });
      }
    );
  }

  /**------------- */

  render() {
    if (this.state.error) {
      return <ErrorComp error={this.state.error} />;
    }

    const wlan0 = this.props.devices.wlan0;

    var scanResults: JSX.Element[] = this.state.WiFiScanResults
      ? this.state.WiFiScanResults.map((res) => {
          const active = wlan0.ActiveConnectionId === res.ssid;
          return <WiFiScanItem
            key={res.ssid}
            name={res.ssid}
            signal={`${res.strength}`}
            wlan0State={active ? wlan0.State : undefined}
            wlan0StateReason={active ? wlan0.State : undefined}
            active={active}
            available={!! wlan0.AvailableConnections.find(c => c.connection.id === res.ssid)}
          />
      })
      : [];

    var wifiStatus = null;
    if (this.state.WiFiInfo) {
      // if (this.state.WiFiInfo.ap_mode) {
      //   wifiStatus = (
      //     <span>
      //       {" "}
      //       Mode: <b>Access Point</b> <MDBIcon icon="broadcast-tower" /> 
      //       <div className="float-right"> SSID:{" "}
      //       <b>
      //         {this.state.WiFiInfo.ssid ? (
      //           this.state.WiFiInfo.ssid
      //         ) : (
      //           <MDBIcon icon="spinner" spin />
      //         )}
      //       </b>
      //       </div>
      //     </span>
      //   );
      // } else {
        // wifiStatus = (
        //   <span>
        //     {" "}
        //     Mode: <b>WiFi client</b> <MDBIcon icon="wifi" />
            
        //     <div className="float-right"> Network: {" "}
        //     <b>
        //       {this.state.WiFiInfo.ssid ? (
        //         this.state.WiFiInfo.ssid
        //       ) : (
        //         <MDBIcon icon="spinner" spin />
        //       )}
        //     </b>{" "}
        //     (
        //     {this.state.WiFiInfo.ip ? (
        //       this.state.WiFiInfo.ip
        //     ) : (
        //       <MDBIcon icon="spinner" spin />
        //     )}
        //     ){"  "}
        //     <span title={this.state.WiFiInfo.state}>
        //       {this.state.WiFiInfo.state ? (
        //         this.state.WiFiInfo.state == "COMPLETED" ? (
        //           <MDBIcon fas icon="check-circle" />
        //         ) : (
        //           <span>
        //             <MDBIcon icon="spinner" spin /> {this.state.WiFiInfo.state}
        //           </span>
        //         )
        //       ) : (
        //         "..."
        //       )}
        //     </span>
        //     </div>
        //   </span>
        // );
      // }
    }

    //nav-justified

    return (
      <MDBContainer className="mt-3">
        <MDBNav tabs className="nav md-pills nav-pills ">
          <MDBNavItem>
            {/* <MDBNavLink
              to="#"
              active={this.state.activeItem === "wifi"}
              onClick={this.toggle("wifi")}
              role="tab"
              // className="bg-info"
              activeClassName="active-link"
            >
              <MDBIcon icon="wifi" /> WiFi
            </MDBNavLink> */}
          </MDBNavItem>
          {/* <MDBNavItem>
						<MDBNavLink
							to="#"
							active={this.state.activeItem === "2"}
							onClick={this.toggle("2")}
							role="tab"
						>
							<MDBIcon icon="signal" /> LTE
						</MDBNavLink>
					</MDBNavItem>
					<MDBNavItem>
						<MDBNavLink
							to="#"
							active={this.state.activeItem === "3"}
							onClick={this.toggle("3")}
							role="tab"
						>
							<MDBIcon icon="envelope" /> Contact
						</MDBNavLink>
					</MDBNavItem> */}
        </MDBNav>
        <MDBTabContent className="card p-2" activeItem={this.state.activeItem}>
          <MDBTabPane tabId="wifi" role="tabpanel">
            <MDBAlert color="info" className="text-justify">
              {/* {wifiStatus ? (
                wifiStatus
              ) : (
                <span>
                  Loading <MDBIcon icon="spinner" spin />{" "}
                </span>
              )} */}
              { nameForState(this.props.devices.wlan0?.State) }
              { this.props.devices.wlan0?.ActiveConnectionId === "WAZIGATE-AP" && " - Access Point Mode"}
            </MDBAlert>

            <MDBAlert color="light">
              Please select a network:
            </MDBAlert>

            <MDBListGroup>
              {scanResults}
              <WiFiScanItem name="Connect to a hidden WiFi" empty signal="0" available={false}/>
            </MDBListGroup>
            <MDBAlert color="light">
              {this.state.scanLoading ? (
                <div style={{textAlign: "center"}} >Checking for available networks <LoadingSpinner
                  type="grow-sm"
                  class="color-light-text-primary"
                /></div> 
              ) : (
                ""
              )}
            </MDBAlert>
          </MDBTabPane>

          {/* <MDBTabPane tabId="2" role="tabpanel">
						<p className="mt-2"></p>
					</MDBTabPane>
					<MDBTabPane tabId="3" role="tabpanel">
						<p className="mt-2"></p>
					</MDBTabPane> */}
        </MDBTabContent>
      </MDBContainer>
    );
  }
}

var stateNames: Record<string, string> = {
  "NmDeviceStateDisconnected": "Disconnecting old connection ...",
  "NmDeviceStatePrepare": "Preparing connection ...",
  "NmDeviceStateConfig": "Preparing connection configuration ...",
  "NmDeviceStateIpConfig": "Preparing connection IP configuration ...",
  "NmDeviceStateIpCheck": "Checking connection IP configuration ...",
  "NmDeviceStateActivated": "Connection activated.",
  "NmDeviceStateDeactivating": "Deactivating connection ...",
  "NmDeviceStateUnavailable": "Device unavailable.",
  "NmDeviceStateUnmanaged": "Device not managed.",
  "NmDeviceStateUnknown": "Device state unknown.",
  "NmDeviceStateSecondaries": "Waiting for connection ...",
  "NmDeviceStateNeedAuth": "Connection requires authorization.",
}

function nameForState(state: string) {
  if(!state) return "";
  return stateNames[state] || ("State: "+state);
}

export default PagesInternet;

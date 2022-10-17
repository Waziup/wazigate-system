import { MDBAlert, MDBBtn, MDBCol, MDBContainer, MDBIcon, MDBRow } from "mdbreact";
import * as React from "react";
import Modal from 'react-bootstrap/Modal';
import { Device, Devices, getBlackout, getNetworkDevices, getVersion, getBuildNr, getWlanDevice, reboot, shutdown } from "../../api";
// import * as API from "../../api";
import ErrorComp from "../Error";
import Clock from "./Clock/Clock";
import * as API from "../../api";
import SensorItem from "./SensorItem";

declare function Notify(msg: string): any;





export interface Props {
  devices: API.Devices
}
export interface State {
  allSensors: any;
  blackout: boolean;
  error: any;
  WiFiLoading: boolean;
  modal: {
    visible: boolean;
    title: string;
    msg: string;
    func: string;
  };
  shutdownLoading: boolean;
  rebootLoading: boolean;
  version: string;
  buildNr: string;
}

class PagesOverview extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      allSensors: null,
      blackout: null,
      WiFiLoading: true,
      error: null,
      modal: {
        visible: false,
        title: "",
        msg: "",
        func: "",
      },
      shutdownLoading: false,
      rebootLoading: false,
      version: "",
      buildNr: "",
    };
  }

  /**------------- */
  _isMounted = false;
  componentDidMount() {
    this._isMounted = true;

    /*-------*/

    // API.getAllSensors().then(
    // 	res => {
    // 		this.setState({
    // 			allSensors: res,
    // 			error: null
    // 		});
    // 	},
    // 	error => {
    // 		this.setState({
    // 			allSensors: null,
    // 			error: error
    // 		});
    // 	}
    // );

    getVersion().then((version) => {
      this.setState({version});
    });

    getBuildNr().then((buildNr) => {
      this.setState({buildNr});
    });

    getBlackout().then(
      (res) => {
        this.setState({
          blackout: res,
        });
      },
      (error) => {
        this.setState({
          blackout: null,
        });
      }
    );

    // if( !this._isMounted) return;
  }
  componentWillUnmount() {
    this._isMounted = false;
  }

  /**------------- */

  shutdown() {
    if (!this._isMounted) return;

    this.setState({
      shutdownLoading: true,
    });

    shutdown().then(
      (res) => {
        Notify(res);
      },
      (error) => {
        console.log(error);
        this.setState({
          shutdownLoading: false,
        });
      }
    );

    this.componentWillUnmount();
  }

  /**------------- */

  reboot() {
    if (!this._isMounted) return;

    this.setState({
      rebootLoading: true,
    });

    reboot().then(
      (res) => {
        Notify(res);
      },
      (error) => {
        console.log(error);
        this.setState({
          rebootLoading: false,
        });
      }
    );
    this.componentWillUnmount();
  }
  /**------------- */

  showModal(title: string, msg: string, func: string) {
    this.setState({
      modal: {
        visible: true,
        title: title,
        msg: msg,
        func: func,
      },
    });
  }

  /**------------- */

  modalClick = () => {
    switch (this.state.modal.func) {
      case "reboot":
        this.reboot();
        break;
      case "shutdown":
        this.shutdown();
        break;
      default:
        console.log("No function found: ", this.state.modal.func);
    }
    this.toggleModal();
  };

  /**------------- */

  toggleModal = () => {
    this.setState({
      modal: {
        visible: !this.state.modal.visible,
        title: this.state.modal.title,
        msg: this.state.modal.msg,
        func: this.state.modal.func,
      },
    });
  };

  /**------------- */

  render() {

    if (this.state.shutdownLoading || this.state.rebootLoading) {
      return <div style={{ marginTop: "20%", textAlign: "center", border: "1px solid #BBB", borderRadius: "5px", padding: "5%", marginLeft: "10%", marginRight: "10%", backgroundColor: "#EEE" }}>
        <h1>Wazigate is not accessible...</h1>
      </div>
    }

    if (this.state.error) {
      return <ErrorComp error={this.state.error} />;
    }

    var wifiStatus = null;
    const wlan0 = this.props.devices.wlan0;
    const eth0 = this.props.devices.eth0;

    if (wlan0) {

      const apConn = wlan0.AvailableConnections.find(conn => conn.connection.id === "WAZIGATE-AP");

      if (wlan0.ActiveConnectionId === "WAZIGATE-AP") {
        wifiStatus = (
          <span>
            <MDBAlert color="info">
              Mode:{" "}
              <b>
                Access Point <MDBIcon icon="broadcast-tower" />
              </b>
            </MDBAlert>
            <MDBAlert color="info">
              SSID:{" "}
              <b>{atob(apConn["802-11-wireless"].ssid)}</b>
            </MDBAlert>
          </span>
        );
      } else {
        wifiStatus = (
          <span>
            <MDBAlert color="info">
              Mode:{" "}
              <b>
                WiFi Client <MDBIcon icon="wifi" />
              </b>
              {/* {"  "}
              <span title={wlan0.State}>
                {wlan0.State ? (
                  wlan0.State == "COMPLETED" ? (
                    <MDBIcon fas icon="check-circle" />
                  ) : (
                    <span>
                      <MDBIcon icon="spinner" spin />{" "}
                      {wlan0.State}
                    </span>
                  )
                ) : (
                  "..."
                )}
              </span> */}
            </MDBAlert>
            <MDBAlert color="info">
              Connected to{" "}
              <b>
                {wlan0.ActiveConnectionId ? (
                  wlan0.ActiveConnectionId
                ) : (
                  <MDBIcon icon="spinner" spin />
                )}
              </b>
            </MDBAlert>
            <MDBAlert color="info">
              IP address:{" "}
              <b>
                {wlan0?.IP4Config ? (
                  wlan0.IP4Config.Addresses[0].Address
                ) : (
                  <MDBIcon icon="spinner" spin />
                )}
              </b>
            </MDBAlert>
          </span>
        );
      }
    }

    return (
      <MDBContainer>
        <MDBRow>
          <MDBCol>
            <div className="card mb-3 mt-3 m-l3 mb-3">
              <h4 className="card-header">
                {" "}
                <MDBIcon
                  spin={!eth0}
                  icon={eth0 ? "network-wired" : "cog"}
                />{" "}
                Ethernet Network
              </h4>
              <div className="card-body">
                <MDBAlert color="info">
                  IP address :{" "}
                  <b>
                    {eth0?.IP4Config ? (
                      eth0.IP4Config.Addresses[0].Address
                    ) : (
                      <MDBIcon icon="spinner" spin />
                    )}
                  </b>
                </MDBAlert>
                {/* <MDBAlert color="info">
                  MAC address :{" "}
                  <b>
                    {eth0.IP4Config ? (
                      eth0.
                    ) : (
                      <MDBIcon icon="spinner" spin />
                    )}
                  </b>
                </MDBAlert> */}
                {/* <MDBAlert color="info">
                  Device :{" "}
                  <b>
                    {this.state.networkDevices ? (
                      this.state.networkDevices.dev
                    ) : (
                      <MDBIcon icon="spinner" spin />
                    )}
                  </b>
                </MDBAlert> */}
              </div>
            </div>

            <div className="card mb-3 mt-3 m-l3 mb-3">
              <h4 className="card-header">
                <MDBIcon far icon="clock" />{" "}
                <span title="Click to change the Timezone">
                  Gateway Clock
                </span>
              </h4>
              <div className="card-body h-100">
                <Clock />
              </div>
            </div>
          </MDBCol>

          { }

          <MDBCol>
            <div className="card mb-3 mt-3 m-l3 mb-3">
              <h4 className="card-header">
                {" "}
                <MDBIcon
                  spin={false}
                  icon={this.state.WiFiLoading ? "cog" : "wifi"}
                />{" "}
                <a href="#/internet" title="Connect to a WiFi network">
                  WiFi Network
                </a>
              </h4>
              <div className="card-body">{wifiStatus}</div>
            </div>
          </MDBCol>
        </MDBRow>

        { }

        <MDBRow>


          <MDBCol>
            <div className="card mb-3 mt-3 m-l3 mb-3">
              <h4 className="card-header">
                <MDBIcon icon="bolt" /> Blackout Protection
              </h4>
              <div className="card-body h-100">
                {this.state.blackout === null ? (
                  <MDBIcon icon="cog" spin />
                ) : this.state.blackout ? (
                  <span>
                    <MDBIcon className="green-text" icon="check-circle" />{" "}
                    Activated
                  </span>
                ) : (
                  <span>
                    <MDBIcon color="orange-text" icon="exclamation-circle" />{" "}
                    Not available
                  </span>
                )}
              </div>
            </div>

            {/* -------------------------- */}

            <div className="card mb-3 mt-3 m-l3 mb-3">
              <h4 className="card-header">
                <MDBIcon icon="power-off" /> Gateway Shutdown
              </h4>
              <div className="card-body">
                <MDBBtn
                  disabled={this.state.shutdownLoading}
                  onClick={() =>
                    this.showModal(
                      "Shutdown the Wazigate",
                      "Are you sure that you want to shutdown the gateway?",
                      "shutdown"
                    )
                  }
                >
                  <MDBIcon
                    icon={this.state.shutdownLoading ? "cog" : "power-off"}
                    className="ml-2"
                    size="1x"
                    spin={this.state.shutdownLoading}
                  />{" "}
                  Shutdown
                </MDBBtn>

                <MDBBtn
                  disabled={this.state.rebootLoading}
                  onClick={() =>
                    this.showModal(
                      "Restart the Wazigate",
                      "Are you sure that you want to restart the gateway?",
                      "reboot"
                    )
                  }
                >
                  <MDBIcon
                    icon={this.state.rebootLoading ? "cog" : "redo"}
                    className="ml-2"
                    size="1x"
                    spin={this.state.rebootLoading}
                  />{" "}
                  Restart
                </MDBBtn>
              </div>
            </div>
          </MDBCol>
          <MDBCol>
            <div className="card mb-3 mt-3 m-l3 mb-3">
              <h4 className="card-header">
                {" "}
                <MDBIcon
                  spin={false}
                  icon={"code-branch"}
                />{" "}
                <span title="Shows the currently installed version of the WaziGate">
                  WaziGate Version
                </span>
              </h4>
              <div className="card-body">{this.state.version + " Buildnumber: " + this.state.buildNr}</div>
            </div>
          </MDBCol>

          { }

          { }
        </MDBRow>

        <MDBRow>
          <MDBCol></MDBCol>
          <MDBCol></MDBCol>
        </MDBRow>

        <Modal show={this.state.modal.visible} onHide={this.toggleModal}>
          <Modal.Header closeButton>
            <Modal.Title>{this.state.modal.title}</Modal.Title>
          </Modal.Header>
          <Modal.Body>{this.state.modal.msg}</Modal.Body>
          <Modal.Footer>
            <MDBBtn onClick={this.toggleModal}>
              No
            </MDBBtn>
            <MDBBtn color="danger" onClick={this.modalClick}>
              Yes
            </MDBBtn>
          </Modal.Footer>
        </Modal>

      </MDBContainer>
    );
  }
}

export default PagesOverview;
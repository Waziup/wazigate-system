import * as React from "react";
import * as API from "../api";
import { Route, Link, HashRouter as Router, Routes } from "react-router-dom";
import MenuBar from "./MenuBar";

import PagesOverview from "./Pages/Overview";
import PagesConfig from "./Pages/Config";
import PagesInternet from "./Pages/Wifi/Internet";
import PagesResources from "./Pages/Resources";
import PagesContainers from "./Pages/Containers/Containers";
import PagesLogs from "./Pages/Logs";
import PagesUpdate from "./Pages/Update";

import "../style/app.scss";
import "@fortawesome/fontawesome-free/css/all.min.css";
import "bootstrap-css-only/css/bootstrap.min.css";
import "mdbreact/dist/css/mdb.css";
import Notifications from "./Notifications";
import { NetworkContext } from "./network-context";

export interface AppCompState {
  devices: API.Devices | null,
  err: unknown,
}

export interface AppCompProps {
  hideLoader: Function;
  showLoader: Function;
}

export type EventDeviceStateChanged = {
  device: string,
  oldState: string,
  newState: string,
  reason: string,
  activeConnectionId: string,
  activeConnectionUUID: string
}

const topicNetworkManagerDeviceEvents = "waziup.wazigate-system/network-manager/device/+"

class AppComp extends React.Component<AppCompProps, AppCompState> {
  constructor(props: AppCompProps) {
    super(props);
    this.state = {
      devices: {},
      err: null,
    }
    this.handleNetworkManagerDeviceEvent = this.handleNetworkManagerDeviceEvent.bind(this);
  }

  //

  componentDidMount() {
    this.props.hideLoader();
    wazigate.subscribe(topicNetworkManagerDeviceEvents, this.handleNetworkManagerDeviceEvent);
    
    API.getNetworkDevices()
      .then(devices => this.setState({devices}))
      .catch(err => this.setState({err}));
  }

  componentWillUnmount() {
    wazigate.unsubscribe(topicNetworkManagerDeviceEvents, this.handleNetworkManagerDeviceEvent);
  }

  handleNetworkManagerDeviceEvent(ev: EventDeviceStateChanged) {
    this.setState(state => {
      if(state.devices && ev.device in state.devices) {
        const device = state.devices[ev.device];
        device.State = ev.newState;
        device.stateReason = ev.reason;
        if (ev.activeConnectionId) {
          device.ActiveConnectionId = ev.activeConnectionId
          device.ActiveConnectionUUID = ev.activeConnectionUUID
        }
        return ({
          devices: {
            ...state.devices,
            [ev.device]: device
          }
        })
      }
    })
  }


  //

  render() {
    return (
      <NetworkContext.Provider value={this.state.devices}>
        <Router>
          <React.Fragment>
            <MenuBar />
            {/* <Route path="/:active?" component={MenuBar} /> */}
            <Routes>
              <Route path="" element={<PagesOverview devices={this.state.devices} />} />
              <Route path="overview" element={<PagesOverview devices={this.state.devices} />} />
              <Route path="config" element={<PagesConfig devices={this.state.devices} />} />
              <Route path="internet" element={<PagesInternet devices={this.state.devices} />} />
              <Route path="resources" element={<PagesResources />} />
              <Route path="containers" element={<PagesContainers />} />
              <Route path="logs" element={<PagesLogs />} />
              <Route path="update" element={<PagesUpdate />} />
              {/* <Route path="/test" element={<PagesTest />} /> */}

              {/* <Route component={Notfound} />   */}
            </Routes>
            <Notifications />
          </React.Fragment>
        </Router>
      </NetworkContext.Provider>
    );
  }
}

export default AppComp;

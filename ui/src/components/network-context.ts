import React from "react";
import { Device } from "../api";

// export type DeviceState = {
//     newState: string,
//     oldState: string,
//     reason: string,
//     connId: string,
// };

export const NetworkContext = React.createContext<Record<string, Device>>({});
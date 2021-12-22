package nm

//go:generate stringer -type=DeviceStateReason -trimprefix=DeviceStateReason
type DeviceStateReason uint32

const (
	DeviceStateReasonNone                        DeviceStateReason = iota // No reason given
	DeviceStateReasonUnknown                                              // Unknown error
	DeviceStateReasonNowManaged                                           // Device is now managed
	DeviceStateReasonNowUnmanaged                                         // Device is now unmanaged
	DeviceStateReasonConfigFailed                                         // The device could not be readied for configuration
	DeviceStateReasonIpConfigUnavailable                                  // IP configuration could not be reserved (no available address, timeout, etc)
	DeviceStateReasonIpConfigExpired                                      // The IP config is no longer valid
	DeviceStateReasonNoSecrets                                            // Secrets were required, but not provided
	DeviceStateReasonSupplicantDisconnect                                 // 802.1x supplicant disconnected
	DeviceStateReasonSupplicantConfigFailed                               // 802.1x supplicant configuration failed
	DeviceStateReasonSupplicantFailed                                     // 802.1x supplicant failed
	DeviceStateReasonSupplicantTimeout                                    // 802.1x supplicant took too long to authenticate
	DeviceStateReasonPppStartFailed                                       // PPP service failed to start
	DeviceStateReasonPppDisconnect                                        // PPP service disconnected
	DeviceStateReasonPppFailed                                            // PPP failed
	DeviceStateReasonDhcpStartFailed                                      // DHCP client failed to start
	DeviceStateReasonDhcpError                                            // DHCP client error
	DeviceStateReasonDhcpFailed                                           // DHCP client failed
	DeviceStateReasonSharedStartFailed                                    // Shared connection service failed to start
	DeviceStateReasonSharedFailed                                         // Shared connection service failed
	DeviceStateReasonAutoipStartFailed                                    // AutoIP service failed to start
	DeviceStateReasonAutoipError                                          // AutoIP service error
	DeviceStateReasonAutoipFailed                                         // AutoIP service failed
	DeviceStateReasonModemBusy                                            // The line is busy
	DeviceStateReasonModemNoDialTone                                      // No dial tone
	DeviceStateReasonModemNoCarrier                                       // No carrier could be established
	DeviceStateReasonModemDialTimeout                                     // The dialing request timed out
	DeviceStateReasonModemDialFailed                                      // The dialing attempt failed
	DeviceStateReasonModemInitFailed                                      // Modem initialization failed
	DeviceStateReasonGsmApnFailed                                         // Failed to select the specified APN
	DeviceStateReasonGsmRegistrationNotSearching                          // Not searching for networks
	DeviceStateReasonGsmRegistrationDenied                                // Network registration denied
	DeviceStateReasonGsmRegistrationTimeout                               // Network registration timed out
	DeviceStateReasonGsmRegistrationFailed                                // Failed to register with the requested network
	DeviceStateReasonGsmPinCheckFailed                                    // PIN check failed
	DeviceStateReasonFirmwareMissing                                      // Necessary firmware for the device may be missing
	DeviceStateReasonRemoved                                              // The device was removed
	DeviceStateReasonSleeping                                             // NetworkManager went to sleep
	DeviceStateReasonConnectionRemoved                                    // The device's active connection disappeared
	DeviceStateReasonUserRequested                                        // Device disconnected by user or client
	DeviceStateReasonCarrier                                              // Carrier/link changed
	DeviceStateReasonConnectionAssumed                                    // The device's existing connection was assumed
	DeviceStateReasonSupplicantAvailable                                  // The supplicant is now available
	DeviceStateReasonModemNotFound                                        // The modem could not be found
	DeviceStateReasonBtFailed                                             // The Bluetooth connection failed or timed out
	DeviceStateReasonGsmSimNotInserted                                    // GSM Modem's SIM Card not inserted
	DeviceStateReasonGsmSimPinRequired                                    // GSM Modem's SIM Pin required
	DeviceStateReasonGsmSimPukRequired                                    // GSM Modem's SIM Puk required
	DeviceStateReasonGsmSimWrong                                          // GSM Modem's SIM wrong
	DeviceStateReasonInfinibandMode                                       // InfiniBand device does not support connected mode
	DeviceStateReasonDependencyFailed                                     // A dependency of the connection failed
	DeviceStateReasonBr2684Failed                                         // Problem with the RFC 2684 Ethernet over ADSL bridge
	DeviceStateReasonModemManagerUnavailable                              // ModemManager not running
	DeviceStateReasonSsidNotFound                                         // The Wi-Fi network could not be found
	DeviceStateReasonSecondaryConnectionFailed                            // A secondary connection of the base connection failed
	DeviceStateReasonDcbFcoeFailed                                        // DCB or FCoE setup failed
	DeviceStateReasonTeamdControlFailed                                   // teamd control failed
	DeviceStateReasonModemFailed                                          // Modem failed or no longer available
	DeviceStateReasonModemAvailable                                       // Modem now ready and available
	DeviceStateReasonSimPinIncorrect                                      // SIM PIN was incorrect
	DeviceStateReasonNewActivation                                        // New connection activation was enqueued
	DeviceStateReasonParentChanged                                        // the device's parent changed
	DeviceStateReasonParentManagedChanged                                 // the device parent's management changed
	DeviceStateReasonOvsdbFailed                                          // problem communicating with Open vSwitch database
	DeviceStateReasonIpAddressDuplicate                                   // a duplicate IP address was detected
	DeviceStateReasonIpMethodUnsupported                                  // The selected IP method is not supported
	DeviceStateReasonSriovConfigurationFailed                             // configuration of SR-IOV parameters failed
	DeviceStateReasonPeerNotFound                                         // The Wi-Fi P2P peer could not be found
)

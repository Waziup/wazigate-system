# wazigate-system

The list of APIs can be found here: https://waziup.github.io/wazigate-system/

A useful link: https://github.com/Waziup/WaziGate/blob/master/docs/System.md


# Status LED Indicator

The recent version of WaziGate has two status LEDs on board and `wazigate-system` indicates the connectivity status via those LEDs.


![LED indicators](assets/LEDs.gif "LED indicators")

## LED 1
`LED 1` indicates the status of Internet connectivity and has two states:

- **Internet connectivity is ok**: it stays on: `_______________________________`
- **No Internet**: it blinks fast like this: `.............`

## LED 2
`LED 2` indicates the WiFi status and it has 3 states:

- **Access Point Mode**: this LED blinks slowly once a second: `__  __  __  __  __`
- **WiFi Client Mode Connected to a router**: it stays on: `_______________________________`
- **Not Connected**: it blinks fast like this: `.............`


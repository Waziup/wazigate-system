# wazigate-system

WaziGate-System is a part of the [Waziup gateway](https://github.com/Waziup/waziup-gateway).

## Development & testing

LoRa inputs can be simulated in the following way.
First log in the container:
```
$ docker exec -it wazigate-system bash
```
Then perform this command:
```
$ python /app/data_acq/edgeCall.py "TC/26.00/HU/62.21/UID/0ff8373bd" "1,16,118,1,19,7,-40" "125,5,12" "2019-06-06T13:16:31.670" "B827EB4E30A8"
```


package wazigate

import (
	"fmt"

	"github.com/Waziup/wazigate-edge/mqtt"
)

func Publish(topic string, data []byte) error {
	client, err := mqtt.Dial("wazigate-edge:1883", "wazigate-system", true, nil, nil)
	if err != nil {
		return err
	}
	if err := client.Publish(&mqtt.Message{
		Topic: topic,
		QoS:   1,
		Data:  data,
	}); err != nil {
		client.Disconnect()
		return err
	}
	pkt, _, err := client.Packet()
	if err != nil {
		client.Disconnect()
		return err
	}
	_, ok := pkt.(*mqtt.PubAckPacket)
	if !ok {
		client.Disconnect()
		return fmt.Errorf("received unknown packet: %s", pkt)
	}
	return client.Disconnect()
}

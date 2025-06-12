package crack

import (
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttCracker struct {
	CrackBase
}

func (s *MqttCracker) Ping() (succ bool, err error) {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", s.Target))
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		err = token.Error()
		if strings.Contains(strings.ToLower(err.Error()), "network error") {
			return false, ERR_CONNECTION
		}
		return false, err
	}
	defer client.Disconnect(250)
	return true, nil
}

func (s *MqttCracker) Crack() (succ bool, err error) {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", s.Target)).SetUsername(s.User).SetPassword(s.Pass)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return false, token.Error()
	}
	defer client.Disconnect(250)
	return true, nil
}

func (*MqttCracker) Class() string {
	return CLASS_MQ_MIDDLEWARE
}

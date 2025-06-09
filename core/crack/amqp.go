package crack

import (
	"fmt"

	"github.com/streadway/amqp"
)

type AmqpCracker struct {
	CrackBase
}

func (s *AmqpCracker) Ping() (succ bool, err error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", "guest", "guest", s.Target))
	if err != nil {
		if amqpErr, ok := err.(*amqp.Error); ok {
			if amqpErr.Code == 403 {
				return false, err
			}
		}
		return false, ERR_CONNECTION
	}
	defer conn.Close()
	return true, nil
}

func (s *AmqpCracker) Crack() (succ bool, err error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", s.User, s.Pass, s.Target))
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return true, nil
}

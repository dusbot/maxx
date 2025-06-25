package ping

import (
	"fmt"
	"testing"
	"time"
)

func TestPing(t *testing.T) {

}

func TestArping(t *testing.T) {
	mac, device, err := TryArping("10.1.30.50")
	if err != nil {
		panic(err)
	}
	fmt.Println(mac, device)
}

func TestTCPPing(t *testing.T) {
	alive, rtt, err := tcpPing("10.1.2.1", 22, "eno1", time.Second*3)
	if err != nil {
		panic(err)
	}
	fmt.Println(alive, rtt)
}

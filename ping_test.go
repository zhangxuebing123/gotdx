package gotdx

import (
	"fmt"
	"testing"
)

func TestPing_Ping(t *testing.T) {
	ping, err := Run("127.0.0.1", 8, Data)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ping.Close()
	ping.Ping(5)
	fmt.Println(ping.PingCount(6))
}

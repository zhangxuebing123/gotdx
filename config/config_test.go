package config

import (
	"fmt"
	"testing"
)

func TestTDXConfig_Load(t *testing.T) {
	var c TDXConfig
	c.Load()
	fmt.Println(c.Remoter(HQHOST))
	fmt.Println(c.Remoter(EXHQHOST))
	fmt.Println(c.Remoter(""))
}
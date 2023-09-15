package edgeturn

import (
	"fmt"
	"testing"
)

func TestPort(t *testing.T) {
	port ,err := GetFreeUDPPort(60000,65535)
	if err != nil {
		fmt.Printf("%s", err.Error())
		t.Fail()
	}
	fmt.Printf("%d", port)
}
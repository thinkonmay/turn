package edgeturn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/thinkonmay/thinkshare-daemon/credential"
)

type TurnCred struct {
	Username string `json:"username"`
	Password string `json:"credential"`
}
type TurnAccount struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Turn     TurnCred `json:"turn_cred"`
}
type TurnInfo struct {
	PublicIP  string `json:"public_ip"`
	PrivateIP string `json:"private_ip"`
	Port      int    `json:"turn_port"`
}

func GetFreeUDPPort(min int,max int) (int, error) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	port := l.LocalAddr().(*net.UDPAddr).Port
	if port > max {
		return 0,fmt.Errorf("invalid port %d",port)
	} else if port < min {
		return GetFreeUDPPort(min,max)
	}
	return port, nil
}
func SetupTurnAccount(proxy credential.Account,
					 min int, 
					 max int) (
					 cred credential.Account,
					 turn TurnCred,
					 info TurnInfo,
					 err error) {
	port,_ := GetFreeUDPPort(min,max)
	info = TurnInfo{
		PublicIP: credential.Addresses.PublicIP,
		PrivateIP: credential.Addresses.PrivateIP,
		Port: port,
	}
	b, _ := json.Marshal(info)

	req, err := http.NewRequest("POST", 
		credential.Secrets.EdgeFunctions.TurnRegister, 
		bytes.NewBuffer(b))
	if err != nil {
		return credential.Account{},TurnCred{},TurnInfo{}, err
	}

	req.Header.Set("username", *proxy.Username)
	req.Header.Set("password", *proxy.Password)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", credential.Secrets.Secret.Anon))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return credential.Account{},TurnCred{},TurnInfo{}, err
	}

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		body_str := string(body)
		return credential.Account{},TurnCred{},TurnInfo{}, fmt.Errorf("response code %d: %s", resp.StatusCode, body_str)
	}

	turn_account := TurnAccount{}
	if err := json.Unmarshal(body, &turn_account); err != nil {
		return credential.Account{},TurnCred{},TurnInfo{}, err
	}

	return credential.Account{
		Username: &turn_account.Username,
		Password: &turn_account.Password,
	},turn_account.Turn,info, nil
}


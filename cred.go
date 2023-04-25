package edgeturn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	SecretDir       = "./secret"
	ProxySecretFile = "./secret/proxy.json"
	UserSecretFile  = "./secret/user.json"
	ConfigFile      = "./secret/config.json"
)

type ApiKey struct {
	Key     string `json:"key"`
	Project string `json:"project"`
}
type Account struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Project  string `json:"project"`
}
type Address struct {
	PublicIP  string `json:"public_ip"`
	PrivateIP string `json:"private_ip"`
	Port      int    `json:"turn_port"`
}

type Secret struct {
	EdgeFunctions struct {
		UserKeygen              string `json:"user_keygen"`
		ProxyRegister           string `json:"proxy_register"`
		SessionAuthenticate     string `json:"session_authenticate"`
		SignalingAuthenticate   string `json:"signaling_authenticate"`
		TurnRegister            string `json:"turn_register"`
		WorkerProfileFetch      string `json:"worker_profile_fetch"`
		WorkerRegister          string `json:"worker_register"`
		WorkerSessionCreate     string `json:"worker_session_create"`
		WorkerSessionDeactivate string `json:"worker_session_deactivate"`
	} `json:"edge_functions"`

	Secret struct {
		Anon string `json:"anon"`
		Url  string `json:"url"`
	} `json:"secret"`

	Google struct {
		ClientId string `json:"client_id"`
	} `json:"google"`

	Conductor struct {
		Hostname string `json:"host"`
		GrpcPort int    `json:"grpc_port"`
	} `json:"conductor"`
}

var Secrets *Secret = &Secret{}
var proj string = os.Getenv("PROJECT")
var Addresses *Address = &Address{
	PublicIP:  GetPublicIPCurl(),
	PrivateIP: GetPrivateIP(),
	Port: GetFreePort(),
}

func init() {
	if proj == "" {
		proj = "avmvymkexjarplbxwlnj"
	}
	commitHash, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err == nil {
		fmt.Printf("current commit hash: %s \n", commitHash)
	} else if commitHash == nil {
		fmt.Println("you are not using git, please download git to have auto update")
	} else if strings.Contains(string(commitHash), "fatal") {
		fmt.Println("you did not clone this repo, please use clone")
	}

	os.Mkdir(SecretDir, os.ModeDir)
	secretFile, err := os.OpenFile(ConfigFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer func() {
		defer secretFile.Close()
		bytes, _ := json.MarshalIndent(Secrets, "", "	")
		secretFile.Truncate(0)
		secretFile.WriteAt(bytes, 0)
	}()

	data, _ := io.ReadAll(secretFile)
	err = json.Unmarshal(data, Secrets)

	if err == nil {
		return
	} // avoid fetch if there is already secrets
	resp, err := http.DefaultClient.Post(fmt.Sprintf("https://%s.functions.supabase.co/constant", proj), "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		fmt.Printf("unable to fetch constant from server %s\n", err.Error())
		return
	} else if resp.StatusCode != 200 {
		fmt.Println("unable to fetch constant from server")
		return
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, Secrets)
}

func UseProxyAccount() (account Account, err error) {
	secret_f, err := os.OpenFile(ProxySecretFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return Account{}, err
	}

	bytes, _ := io.ReadAll(secret_f)
	err = json.Unmarshal(bytes, &account)
	if err != nil {
		fmt.Println("none proxy account provided, please provide (look into ./secret folder on the machine you setup proxy account)")
		fmt.Printf("username : ")
		fmt.Scanln(&account.Username)
		fmt.Printf("password : ")
		fmt.Scanln(&account.Password)
		account.Project = proj

		defer func() {
			bytes, _ := json.MarshalIndent(account, "", "	")
			secret_f.Truncate(0)
			secret_f.WriteAt(bytes, 0)
			secret_f.Close()
		}()

		return account, nil
	}

	secret_f.Close()
	return account, nil
}

type TurnAccount struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Turn     TurnCred `json:"turn_cred"`
}
type TurnCred struct {
	Username string `json:"username"`
	Password string `json:"credential"`
}

func SetupTurnAccount(proxy Account) (
	cred Account,
	turn TurnCred,
	err error) {

	b, _ := json.Marshal(Addresses)
	req, err := http.NewRequest("POST", Secrets.EdgeFunctions.TurnRegister, bytes.NewBuffer(b))
	if err != nil {
		return Account{},TurnCred{}, err
	}

	req.Header.Set("username", proxy.Username)
	req.Header.Set("password", proxy.Password)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", Secrets.Secret.Anon))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Account{},TurnCred{}, err
	}

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		body_str := string(body)
		return Account{},TurnCred{}, fmt.Errorf("response code %d: %s", resp.StatusCode, body_str)
	}

	turn_account := TurnAccount{}
	if err := json.Unmarshal(body, &turn_account); err != nil {
		return Account{},TurnCred{}, err
	}

	return Account{
		Username: turn_account.Username,
		Password: turn_account.Password,
	},turn_account.Turn, nil
}

func GetPublicIPCurl() string {
	resp, err := http.Get("https://ifconfig.me/ip")
	if err != nil {
		return ""
	}

	ip := make([]byte, 1000)
	size, err := resp.Body.Read(ip)
	if err != nil {
		return ""
	}

	return string(ip[:size])
}

func GetPrivateIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func GetFreePort() (int) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}



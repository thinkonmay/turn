package main

import (
	"edgeturn"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thinkonmay/thinkshare-daemon/credential"
)

const (
	threadNum = 4
	realm     = "thinkmay.net"
)

var (
	proj 	 = "https://supabase.thinkmay.net"
	anon_key = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ewogICJyb2xlIjogImFub24iLAogICJpc3MiOiAic3VwYWJhc2UiLAogICJpYXQiOiAxNjk0MDE5NjAwLAogICJleHAiOiAxODUxODcyNDAwCn0.EpUhNso-BMFvAJLjYbomIddyFfN--u-zCf0Swj9Ac6E"
)
func init() {
	project := os.Getenv("TM_PROJECT")
	key     := os.Getenv("TM_ANONKEY")
	if project != "" {
		proj = project
	}
	if key != "" {
		anon_key = key
	}
}

func main() {
	credential.SetupEnv(proj,anon_key)
	proxy_cred, err := credential.InputProxyAccount()
	if proxy_cred.Username == nil && proxy_cred.Password == nil {
		Username := os.Getenv("PROXY_USERNAME")
		Password := os.Getenv("PROXY_PASSWORD")
		if Username == "" && Password == "" {
			panic(fmt.Errorf("no proxy account found"))
		}

		proxy_cred = credential.Account{
			Username: &Username,
			Password: &Password,
		}
	}

	fmt.Printf("proxy account found %d, continue\n",proxy_cred)
	worker_cred,turn_cred,info, err := edgeturn.SetupTurnAccount(proxy_cred)
	go func() {
		agent := edgeturn.NewSupabaseAgent(credential.Secrets.Secret.Url,credential.Secrets.Secret.Anon)
		uid,err := agent.SignIn(*worker_cred.Username,*worker_cred.Password)
		if err != nil {
			panic(err)
		}
		for {
			err := agent.Ping(uid)
			if err != nil {
				fmt.Println(err.Error())
			}
			time.Sleep(10 * time.Second)
		}
	}()


	username, password := turn_cred.Username,turn_cred.Password
	if err != nil {
		fmt.Printf("failed to setup worker account: %s", err.Error())
		return
	}

	s,err := edgeturn.SetupTurn(info.PublicIP,username,password, info.Port,65535,60000)
	if err != nil {
		panic(err)
	}

	// Block until user sends SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs


	
	if err = s.Close(); err != nil {
		log.Panic(err)
	}
}

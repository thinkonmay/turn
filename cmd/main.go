package main

import (
	"context"
	"edgeturn"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pion/turn/v2"
	"github.com/thinkonmay/thinkshare-daemon/credential"
	"golang.org/x/sys/unix"
)

const (
	threadNum = 4
	realm     = "thinkmay.net"
)

const (
	proj 	 = "https://supabase.thinkmay.net"
	anon_key = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ewogICJyb2xlIjogImFub24iLAogICJpc3MiOiAic3VwYWJhc2UiLAogICJpYXQiOiAxNjk0MDE5NjAwLAogICJleHAiOiAxODUxODcyNDAwCn0.EpUhNso-BMFvAJLjYbomIddyFfN--u-zCf0Swj9Ac6E"
)

func main() {
	credential.SetupEnv(proj,anon_key)
	proxy_cred, err := credential.InputProxyAccount()
	if err != nil {
		fmt.Printf("failed to find proxy account: %s", err.Error())
		return
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

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", info.Port))
	if err != nil {
		log.Fatalf("Failed to parse server address: %s", err)
	}

	// Cache -users flag for easy lookup later
	// If passwords are stored they should be saved to your DB hashed using turn.GenerateAuthKey
	usersMap := map[string][]byte{}
	usersMap[username] = turn.GenerateAuthKey(username, realm, password)

	// Create `numThreads` UDP listeners to pass into pion/turn
	// pion/turn itself doesn't allocate any UDP sockets, but lets the user pass them in
	// this allows us to add logging, storage or modify inbound/outbound traffic
	// UDP listeners share the same local address:port with setting SO_REUSEPORT and the kernel
	// will load-balance received packets per the IP 5-tuple
	listenerConfig := &net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			var operr error
			if err = conn.Control(func(fd uintptr) {
				operr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			}); err != nil {
				return err
			}

			return operr
		},
	}

	relayAddressGenerator := &turn.RelayAddressGeneratorStatic{
		RelayAddress: net.ParseIP(info.PublicIP), // Claim that we are listening on IP passed by user
		Address:      "0.0.0.0",             // But actually be listening on every interface
	}

	packetConnConfigs := make([]turn.PacketConnConfig, threadNum)
	for i := 0; i < threadNum; i++ {
		conn, listErr := listenerConfig.ListenPacket(context.Background(), addr.Network(), addr.String())
		if listErr != nil {
			log.Fatalf("Failed to allocate UDP listener at %s:%s", addr.Network(), addr.String())
		}

		packetConnConfigs[i] = turn.PacketConnConfig{
			PacketConn:            conn,
			RelayAddressGenerator: relayAddressGenerator,
		}

		log.Printf("Server %d listening on %s\n", i, conn.LocalAddr().String())
	}

	s, err := turn.NewServer(turn.ServerConfig{
		Realm: realm,
		// Set AuthHandler callback
		// This is called every time a user tries to authenticate with the TURN server
		// Return the key for that user, or false when no user is found
		AuthHandler: func(username string, realm string, srcAddr net.Addr) ([]byte, bool) {
			if key, ok := usersMap[username]; ok {
				return key, true
			}
			return nil, false
		},
		// PacketConnConfigs is a list of UDP Listeners and the configuration around them
		PacketConnConfigs: packetConnConfigs,
	})
	if err != nil {
		log.Panicf("Failed to create TURN server: %s", err)
	}

	// Block until user sends SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	if err = s.Close(); err != nil {
		log.Panicf("Failed to close TURN server: %s", err)
	}
}

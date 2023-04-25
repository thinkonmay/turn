package edgeturn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nedpals/supabase-go"
)

type SupabaseAgent struct {
	url string
	anon_key string
}

func NewSupabaseAgent(url string, 
					  key string, 
					  ) (*SupabaseAgent) {
	return &SupabaseAgent{
		url: url,
		anon_key: key,
	}
}


func (agent *SupabaseAgent)	SignIn(username string, 
								   password string,
								   )(uid string,
									err error)  {
	sb := supabase.CreateClient(agent.url,agent.anon_key);
	auth,err := sb.Auth.SignIn(context.Background(),supabase.UserCredentials{
		Email: username,
		Password: password, 
	});

	if err != nil {
		return "",err
	}


	return auth.User.ID,nil
}


func (agent *SupabaseAgent)	Ping( uid string)( err error)  {
	body,_ := json.Marshal( struct {
		AccountID string `json:"account_uid"`
	}{
		AccountID: uid,
	})

	req,err := http.NewRequest("POST",fmt.Sprintf("%s/rest/v1/rpc/ping_account",agent.url),bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp,err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		data,_ := io.ReadAll(resp.Body)
		return fmt.Errorf(string(data))
	}

	return
}
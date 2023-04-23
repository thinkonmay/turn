package edgeturn

import (
	"context"
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
	sb := supabase.CreateClient(agent.url,agent.anon_key);
	_,err = sb.DB.Rpc("ping_account",struct {
		AccountID string `json:"account_uid"`
	}{
		AccountID: uid,
	})
	return
}
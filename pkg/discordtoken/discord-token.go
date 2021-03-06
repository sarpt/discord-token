package discordtoken

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/sarpt/discord-token/pkg/oauth"
	"golang.org/x/oauth2"
)

const authURL string = "https://discordapp.com/api/oauth2/authorize"
const tokenURL string = "https://discordapp.com/api/oauth2/token"
const discordTokenConfigDir string = "discord-token"
const tokenFileName string = "token.json"

// Config manipulates GenerateToken behaviour
type Config struct {
	Client   ClientInfo
	Redirect url.URL
	Scopes   []string
}

// GenerateToken returns new token generated by showing user authorization window and estabilishing connection to Discord authorization and token URLs
func GenerateToken(ctx context.Context, conf Config) (*oauth2.Token, error) {
	config, state := oauth.GetAuthConfig(conf.Client.ID, conf.Client.Secret, conf.Scopes, conf.Redirect, authURL, tokenURL)

	promptUserForAccess(config.AuthCodeURL(state))

	var token *oauth2.Token
	authorization, err := oauth.GetAuthorizedClient(ctx, config, state, conf.Redirect)
	if err != nil {
		return token, err
	}

	return authorization.Token, nil
}

// WriteTokenFile attempts writing token file to path. When path is empty, the XDG_CONFIG_HOME is used to guess correct path
func WriteTokenFile(path string, token oauth2.Token) error {
	var tokenPath string
	var err error

	if path == "" {
		tokenPath, err = getFilePath(tokenFileName)
		if err != nil {
			return err
		}

	} else {
		tokenPath = path
	}

	tokenFile, err := os.Create(tokenPath)
	if err != nil {
		return err
	}

	return oauth.WriteTokenToJSON(tokenFile, token)
}

// GetContext returns context related passed timeout
func GetContext(timeout int) (context.Context, context.CancelFunc) {
	var ctx context.Context
	var cancel context.CancelFunc

	if timeout == 0 {
		ctx = context.Background()
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	}

	return ctx, cancel
}

// GetRedirect returns information about server which will listen during generation of token for the redirect from browser with code and state
func GetRedirect(address string, route string) url.URL {
	return url.URL{
		Host:   address,
		Path:   route,
		Scheme: "http",
	}
}

func promptUserForAccess(authorizationURL string) {
	xdgCommand := exec.Command("xdg-open", authorizationURL)
	xdgCommand.Run()
}

func getFilePath(fileName string) (string, error) {
	confDir, err := os.UserConfigDir()

	return path.Join(confDir, discordTokenConfigDir, fileName), err
}

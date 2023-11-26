package pkg

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var (
	clientID     = os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
)

func Auth() {
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8888/callback",
		Scopes:       []string{"user-read-private", "user-library-read"},
		Endpoint:     spotify.Endpoint,
	}

	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println("Please visit the following URL and authorize the app:", url)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := conf.Exchange(context.Background(), code)
		if err != nil {
			fmt.Fprintf(w, "Error getting token: %s", err)
			return
		}

		// アクセストークンとリフレッシュトークンをファイルに保存
		err = SaveTokens("tokens.txt", token)
		if err != nil {
			fmt.Fprintf(w, "Error saving tokens: %s", err)
			return
		}

		fmt.Fprintf(w, "Authorization successful, tokens saved.")
	})

	http.ListenAndServe(":8888", nil)
}

func SaveTokens(filename string, token *oauth2.Token) error {
	data := fmt.Sprintf("Access Token: %s\nRefresh Token: %s\n", token.AccessToken, token.RefreshToken)
	return ioutil.WriteFile(filename, []byte(data), 0644)
}

func GetToken(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Error opening token file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Access Token:") {
			accessToken := strings.TrimSpace(strings.TrimPrefix(line, "Access Token:"))
			return accessToken, nil
		}
	}

	return "", fmt.Errorf("Access token not found in %s", filePath)
}

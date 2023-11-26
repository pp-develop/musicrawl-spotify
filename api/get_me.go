package api

import (
    "bufio"
    "fmt"
    "net/http"
	"io/ioutil"
    "os"
    "strings"
)

// Spotifyのユーザー情報を取得する関数
func getSpotifyUserProfile(accessToken string) (string, error) {
    req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
    if err != nil {
        return "", err
    }
    req.Header.Set("Authorization", "Bearer "+accessToken)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    response, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(response), nil
}

func getMe() {
    // tokens.txt からアクセストークンを読み込む
    file, err := os.Open("tokens.txt")
    if err != nil {
        fmt.Println("Error opening token file:", err)
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var accessToken string
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "Access Token:") {
            accessToken = strings.TrimSpace(strings.TrimPrefix(line, "Access Token:"))
            break
        }
    }

    if accessToken == "" {
        fmt.Println("Access token not found in tokens.txt")
        return
    }

    // Spotify API を叩く
    userProfile, err := getSpotifyUserProfile(accessToken)
    if err != nil {
        fmt.Println("Error getting Spotify user profile:", err)
        return
    }

    fmt.Println("Spotify User Profile:", userProfile)
}

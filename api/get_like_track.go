package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/pp-develop/musicrawl-spotify/pkg"

	"github.com/go-redis/redis/v8"
)

var (
	ctx = context.Background()
)

type Track struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ArtistID string `json:"artist_id"`
	Genre    string `json:"genre"`
}

// Spotify APIのトラック情報応答を表す
type SpotifyTracksResponse struct {
	Items []struct {
		Track struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Artists []struct {
				ID string `json:"id"`
			} `json:"artists"`
			// 他のフィールドは省略
		} `json:"track"`
	} `json:"items"`
}

// Spotifyのユーザーのお気に入りの曲を取得
func GetFavoriteTracks(accessToken string) ([]Track, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/tracks", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respData SpotifyTracksResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, err
	}

	var tracks []Track
	for _, item := range respData.Items {
		track := Track{
			ID:       item.Track.ID,
			Name:     item.Track.Name,
			ArtistID: item.Track.Artists[0].ID,
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// Redisにトラック情報を保存
func SaveTrackInfoToRedis(client *redis.Client, track Track) error {
	// トラック情報をJSONにシリアライズ
	jsonData, err := json.Marshal(track)
	if err != nil {
		return err
	}

	// JSONデータをRedisに保存
	err = client.Set(ctx, track.ID, jsonData, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func ScanAndPrintAllKeys(client *redis.Client) {
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(ctx, cursor, "*", 10).Result()
		if err != nil {
			fmt.Println("Error scanning keys:", err)
			return
		}

		for _, key := range keys {
			result, err := client.Get(ctx, key).Result()
			if err != nil {
				fmt.Println("Error getting value for key:", key, err)
				continue
			}

			var track Track
			err = json.Unmarshal([]byte(result), &track)
			if err != nil {
				fmt.Println("Error unmarshalling track:", err)
				continue
			}

			// トラック情報をより読みやすく表示
			fmt.Printf("Key: %s, Track ID: %s, Name: %s, Artist ID: %s\n",
				key, track.ID, track.Name, track.ArtistID, track.Genre)
		}

		// 全てのキーをスキャンし終えた場合、ループを抜ける
		if cursor == 0 {
			break
		}
	}
}

func GetLikeTracks() {
	// Redisクライアントの設定
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Redisサーバーのアドレス
		Password: "",           // パスワード（設定されている場合）
		DB:       0,            // 使用するDB
	})

	accessToken, err := pkg.getToken("tokens.txt")
	if err != nil {
		fmt.Println("Error getting token:", err)
		return
	}

	tracks, err := GetFavoriteTracks(accessToken)
	if err != nil {
		fmt.Println("Error getting favorite tracks:", err)
		return
	}

	for _, track := range tracks {
		err := SaveTrackInfoToRedis(redisClient, track)
		if err != nil {
			fmt.Println("Error saving track to Redis:", err)
			continue
		}
	}

	fmt.Println("Tracks information saved to Redis successfully.")

	ScanAndPrintAllKeys(redisClient)
}

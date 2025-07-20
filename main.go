package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type BlizzardAuthToken struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	Sub              string `json:"sub"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type WowApiTokenInfo struct {
	LastUpdatedTimestamp int64 `json:"last_updated_timestamp"`
	Price                int   `json:"price"`
}

type DiscordWebhookApi struct {
	Content string `json:"content"`
}

func main() {
	blzAuthAPIEndpoint := "https://oauth.battle.net/token"
	blzTokenAPIEndpoint := "https://us.api.blizzard.com/data/wow/token/?namespace=dynamic-us"
	blzClientID := os.Getenv("TKNT_BLIZZARD_CLIENTID")
	blzClientSecret := os.Getenv("TKNT_BLIZZARD_CLIENTSECRET")
	discordWebhook := os.Getenv("TKNT_DISCORD_WEBHOOK")
	notificationThresholdStr := os.Getenv("TKNT_NOTIFICATION_THRESHOLD")
	var notificationThreshold int

	if blzClientID == "" {
		log.Fatalln("Environment variable TKNT_BLIZZARD_CLIENTID is not set.")
	}
	if blzClientSecret == "" {
		log.Fatalln("Environment variable TKNT_BLIZZARD_CLIENTSECRET is not set.")
	}
	if discordWebhook == "" {
		log.Fatalln("Environment variable TKNT_DISCORD_WEBHOOK is not set.")
	}
	notificationThreshold, err := strconv.Atoi(notificationThresholdStr)
	if err != nil {
		notificationThreshold = 0
	}

	auth_token := getAuthToken(blzClientID, blzClientSecret, blzAuthAPIEndpoint)

	req, err := http.NewRequest("GET", blzTokenAPIEndpoint, nil)
	if err != nil {
		log.Fatalf("Error creating wow token api request: %v\n", err)
	}
	req.Header.Set("Authorization", "Bearer "+auth_token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending wow token api request: %v\n", err)
	}

	tokenBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading wow token api response body: %v", err)
	}
	resp.Body.Close()

	var token WowApiTokenInfo
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		log.Fatalf("Error unmarshaling json from wow token api response body: %v\n", err)
	}

	fmt.Println("Current token price: ", token.Price/10000)
	fmt.Println("Last updated: ", time.UnixMilli(token.LastUpdatedTimestamp))

	if token.Price >= notificationThreshold {
		sendDiscordWebhook(
			discordWebhook,
			"WoW Tokens are above the notification threshold set at "+
				strconv.Itoa(notificationThreshold/10000)+
				". Current value is "+
				strconv.Itoa(token.Price/10000))
	}

}

func getAuthToken(clientId string, clientSecret string, authEndpoint string) string {
	// This could probably be achieved just by setting the body to "grant_type=client_credentials"
	// and then setting the correct content-type header.
	var body bytes.Buffer
	bodyWriter := multipart.NewWriter(&body)
	bodyWriter.WriteField("grant_type", "client_credentials")
	bodyWriter.Close()

	req, err := http.NewRequest("POST", authEndpoint, &body)
	if err != nil {
		log.Fatalf("Error creating auth token request: %v\n", err)
	}

	req.SetBasicAuth(clientId, clientSecret)
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending auth token request: %v\n", err)
	}
	defer resp.Body.Close()

	resp_body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading body of response from auth token request: %v\n", err)
	}

	var authToken BlizzardAuthToken
	err = json.Unmarshal(resp_body, &authToken)
	if err != nil {
		log.Fatalf("Error unmarshaling auth request response: %v\n", err)
	}

	if authToken.Error != "" {
		log.Fatalf("Authenication API returned an error: %s [%s]", authToken.Error, authToken.ErrorDescription)
	}

	return authToken.AccessToken
}

func sendDiscordWebhook(webhookEndpoint string, message string) {
	whApi := DiscordWebhookApi{
		Content: message,
	}

	whJson, err := json.Marshal(whApi)
	if err != nil {
		log.Println("Error creating discord webhook request: ", err)
		return
	}

	req, err := http.NewRequest("POST", webhookEndpoint, bytes.NewBuffer(whJson))
	if err != nil {
		log.Println("Error creating discord webhook request: ", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending discord webhook request: ", err)
		return
	}
	defer resp.Body.Close()
}

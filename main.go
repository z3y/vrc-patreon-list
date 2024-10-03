package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
)

type PatreonData struct {
	Data []struct {
		Attributes struct {
			FullName      string `json:"full_name"`
			PatreonStatus string `json:"patron_status"`
			Cents         int    `json:"currently_entitled_amount_cents"`
		} `json:"attributes"`
		Relationships struct {
			User struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"user"`
		} `json:"relationships"`
	} `json:"data"`
}

type PatreonInfo struct {
	Name  string `json:"name"`
	Cents int    `json:"cents"`
}

func main() {
	oauthToken := os.Getenv("PATREON_TOKEN")

	if oauthToken == "" {
		log.Fatalln("Empty token")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.patreon.com/api/oauth2/v2/campaigns/3498918/members?page%5Bsize%5D=1000&include=user&fields%5Bmember%5D=full_name,patron_status,pledge_relationship_start,pledge_cadence,currently_entitled_amount_cents", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	req.Header.Add("Authorization", "Bearer "+oauthToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch data. Status Code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(1)
	}

	var patreonData PatreonData
	if err := json.Unmarshal(body, &patreonData); err != nil {
		fmt.Println("Error parsing JSON:", err)
		os.Exit(1)
	}

	//fmt.Printf(string(body))

	sort.Slice(patreonData.Data, func(i, j int) bool {
		return patreonData.Data[i].Attributes.Cents > patreonData.Data[j].Attributes.Cents
	})

	result := []PatreonInfo{}

	for _, member := range patreonData.Data {
		if member.Attributes.PatreonStatus != "active_patron" {
			continue
		}
		//fmt.Printf("User: %s, Cents: %d\n", member.Attributes.FullName, member.Attributes.Cents)

		info := PatreonInfo{Name: member.Attributes.FullName, Cents: member.Attributes.Cents}

		result = append(result, info)
	}

	bytes, err := json.MarshalIndent(result, "", "	")

	if err != nil {
		fmt.Println(err)
	}

	os.WriteFile("patrons.json", bytes, 0644)
}

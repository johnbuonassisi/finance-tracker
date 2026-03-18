package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Categorization Tool
//
// Given a google sheet by id, with a column containing transaction descriptions
// output a corresponding category in another column.
//
// Inputs:
// - credentials file
// - sheet id
// - list of categories
// - column with descriptions
// - column to put categories into
//
// Outputs:
// - column with output categories
//
// Algorithm:
//
// - clean the descriptions
//   - lowercase all letters
//
// - find any known vendors in each description and set the category

const (
	credentialFilePath = "/Users/john/.config/gcloud/finance-tracker-credentials.json"

	spreadsheetID = "1gVUDsKybwI0XOtB4WPDjJvTuYK4DN_-CNmqCDsv7o-E"

	txnSheetName           = "Transactions"
	txnDescriptionColumn   = "C"
	txnCategoryColumn      = "D"
	txnDescriptionStartRow = 2

	categoriesSheetName = "Categories"
	categoriesColumn    = "B"
	categoriesStartRow  = 2
)

func main() {
	ctx := context.Background()
	b, err := os.ReadFile(credentialFilePath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.
		BatchGet(spreadsheetID).
		Ranges("Transactions!C2:C", "Transactions!G2:G", "Rules!A2:A", "Rules!B2:B").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.ValueRanges) != 4 {
		log.Fatalf("Failed to retrive two columns")
	}

	descriptions := resp.ValueRanges[0]
	fmt.Printf("descriptions: %d\n", len(descriptions.Values))
	fmt.Printf("last description %s\n", descriptions.Values[len(descriptions.Values)-1])
	subCategories := resp.ValueRanges[1]
	fmt.Printf("categories: %d\n", len(subCategories.Values))
	fmt.Printf("last category %s\n", subCategories.Values[len(subCategories.Values)-1])

	rulesKeywords := resp.ValueRanges[2].Values
	rulesSubCategorys := resp.ValueRanges[3].Values

	rulesKeywordToSubCategory := map[string]string{}
	for idx, rulesKeyword := range rulesKeywords {
		rulesKeywordToSubCategory[fmt.Sprintf("%s", rulesKeyword[0])] = fmt.Sprintf("%s", rulesSubCategorys[idx][0])
	}
	fmt.Printf("keywordToSubCateogory: %v", rulesKeywordToSubCategory)

	data := []*sheets.ValueRange{}

	var uncategorizedRows int

	if len(descriptions.Values) == 0 {
		log.Fatalf("No data found.")
	}

	// loop through the descriptions
	// uppercase the description
	// if subcategory not specified
	// loop through all the descriptions keyword lookups
	// if a keyword is contained in a description, put a subcategory on the txn
	for idx, descriptions := range descriptions.Values {

		description := strings.ToUpper(fmt.Sprintf("%s", descriptions[0]))
		wasCategorized := false

		if subCategories.Values == nil || idx >= len(subCategories.Values) || len(subCategories.Values[idx]) == 0 {
			for keyword, subCategory := range rulesKeywordToSubCategory {
				//fmt.Printf("%s - %s - %s\n", description, keyword, subCategory)
				if strings.Contains(description, keyword) {
					data = append(data, &sheets.ValueRange{
						Range:  fmt.Sprintf("Transactions!G%d", idx+2),
						Values: [][]any{{fmt.Sprintf("%s", subCategory)}},
					})
					wasCategorized = true
				}
			}
		}
		if !wasCategorized {
			uncategorizedRows++
		}
	}

	fmt.Printf("Found %d rows\n", len(descriptions.Values))
	fmt.Printf("Categorized %d rows\n", len(data))
	fmt.Printf("Uncategorized %d rows\n", uncategorizedRows)

	// write as a batch for non-contiguous cells needs to be done this way so each
	// update specifies a specific cell it is updating
	req := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED", // like a user entered it in google sheets, formatting gets applied etc.
		Data:             data,
	}
	_, err = srv.Spreadsheets.Values.
		BatchUpdate(spreadsheetID, req).
		Do()
	if err != nil {
		log.Fatalf("failed to update, %v", err)
	}
}

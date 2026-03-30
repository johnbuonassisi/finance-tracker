package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/johnbuonassisi/finance-tracker/internal/auth"
)

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

	client, err := auth.NewDefaultClient()
	if err != nil {
		log.Fatalf("Unable to create new Sheets client: %v", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Service: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.
		BatchGet(spreadsheetID).
		Ranges(fmt.Sprintf("%s!C2:C", txnSheetName),
			fmt.Sprintf("%s!G2:G", txnSheetName),
			"Rules!A2:A",
			"Rules!B2:B").
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
						Range:  fmt.Sprintf("%s!G%d", txnSheetName, idx+2),
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

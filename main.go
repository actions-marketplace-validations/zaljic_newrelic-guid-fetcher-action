package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// This struct is used to unmarshal the JSON returned by the New Relic API.
type GraphQL struct {
	Data struct {
		Actor struct {
			EntitySearch struct {
				Count   int    `json:"count"`
				Query   string `json:"query"`
				Results struct {
					Entities []struct {
						EntityType string `json:"entityType"`
						GUID       string `json:"guid"`
						Name       string `json:"name"`
					} `json:"entities"`
				} `json:"results"`
			} `json:"entitySearch"`
		} `json:"actor"`
	} `json:"data"`
}

// This function is the entry point for the action. It is responsible for
// parsing the input parameters, calling the functions that fetch the
// application ID from the New Relic API, and setting the output parameter.
func main() {
	// Get the input parameters from the environment variables.
	newrelicApiKey := os.Getenv("INPUT_NEWRELICAPIKEY")
	newrelicRegion := os.Getenv("INPUT_NEWRELICREGION")
	newrelicAppID := os.Getenv("INPUT_NEWRELICAPPID")

	// Return an error if the newrelicApiKey input parameter is not set.
	if newrelicApiKey == "" {
		fmt.Println("NewRelic API key not specified.")
		os.Exit(1)
	}

	// Return an error if the newrelicAppID input parameter is not set.
	if newrelicAppID == "" {
		fmt.Println("NewRelic app ID not specified.")
		os.Exit(1)
	}

	// Set the NewRelic GraphQL endpoint based on the region specified in the
	// newrelicRegion input parameter.
	newrelicApiEndpoint := ""
	// The New Relic GraphQL endpoint is different for US and EU regions.
	if newrelicRegion == "US" {
		newrelicApiEndpoint = "https://api.newrelic.com/graphql"
	} else if newrelicRegion == "EU" {
		newrelicApiEndpoint = "https://api.eu.newrelic.com/graphql"
		// If the region is not US or EU, exit with an error.
	} else {
		fmt.Println("Invalid NewRelic region specified.")
		os.Exit(1)
	}

	// Call the getGUID function to fetch the list of applications from
	// the NewRelic GraphQL endpoint.
	graphqlResponse, err := getGUID(newrelicApiKey, newrelicApiEndpoint, newrelicAppID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Call the getApplicationGUID function to get the application GUID from the
	// GraphQL response.
	applicationGUID := getApplicationGUID(graphqlResponse)

	// Print the output parameter to stdout.
	fmt.Printf(`::set-output name=appGUID::%s`, applicationGUID)
}

// This function sends a HTTP POST request to the endpoint specified in the
// newrelicApiEndpoint input parameter and returns the GraphQL response
// returned by the NewRelic API. It is assumed that the GraphQL response
// contains a list of applications.
func getGUID(newrelicApiKey string, newrelicApiEndpoint string, newrelicAppID string) (GraphQL, error) {
	// Create a new net/http client.
	client := &http.Client{}

	// Specify data to be sent in the HTTP request body.
	dataString := fmt.Sprintf(`{"query":"{ actor { entitySearch(query: \"domainId=%s\") { count query results { entities { entityType name guid } } } } }\n","variables":null}`, newrelicAppID)
	data := strings.NewReader(dataString)

	// Send a HTTP GET request using net/http to the NewRelic GraphQL endpoint
	// specified in the newrelicApiEndpoint input parameter.
	req, err := http.NewRequest("POST", newrelicApiEndpoint, data)
	if err != nil {
		log.Fatal(err)
	}

	// Set the Api-Key header to the value of the newrelicApiKey input parameter.
	req.Header.Set("Api-Key", newrelicApiKey)

	// Set the Content-Type header to application/x-www-form-urlencoded.
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request using the net/http client.
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// Close the HTTP response body.
	defer resp.Body.Close()

	// Print http status code.
	fmt.Println(resp.StatusCode)

	// Return an error if the HTTP status code is not 200.
	if resp.StatusCode != 200 {
		return GraphQL{}, errors.New("HTTP status code is not 200")
	}

	// Unmarshal the HTTP response body into the GraphQL struct.
	var graphqlResponse GraphQL
	err = json.NewDecoder(resp.Body).Decode(&graphqlResponse)
	if err != nil {
		log.Fatal(err)
	}

	// Return the GraphQL response.
	return graphqlResponse, nil
}

// This function returns the application GUID of the previously fetched
// GraphQL response. It is assumed that the entities list only contains one
// GUID.
func getApplicationGUID(graphqlResponse GraphQL) string {
	// Return the application GUID.
	return graphqlResponse.Data.Actor.EntitySearch.Results.Entities[0].GUID
}

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/twinj/uuid"
)

// Struct for incoming recipt requests given as a JSON.
type Receipt struct {
	Retailer     string  `json:"retailer"`
	Total        float64 `json:"total,string"`
	PurchaseDate string  `json:"purchaseDate"`
	PurchaseTime string  `json:"purchaseTime"`
	Items        []Item  `json:"items,omitempty"`
}

// Struct for list items from receipt processing requests given as JSON.
type Item struct {
	Description string  `json:"shortDescription"`
	Price       float64 `json:"price,string"`
}

// Struct for returning a newly generated receipt id given as JSON.
type ReceiptResponse struct {
	ID string `json:"id"`
}

// Struct for returning the calculated points given a receipt object.
type PointsResponse struct {
	Points int `json:"points"`
}

var receipts = make(map[string]*Receipt)

// Function to handle receipt requests.
func processReceiptsHandler(w http.ResponseWriter, r *http.Request) {

	//Parse given JSON from the request.
	var receipt Receipt
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Generate a unique ID.
	id := uuid.NewV4().String()

	//generate a response JSON body.
	response := ReceiptResponse{ID: id}

	//Store the receipt object in the receipts map using the generated id as the key.
	receipts[id] = &receipt

	//Send the response.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Function to handle points response given a receipt id.
func getPointsHandler(w http.ResponseWriter, r *http.Request) {

	//Parameters for request r.
	params := mux.Vars(r)

	//Extract the id from the request parameters.
	id := params["id"]

	//See if the receipt exists in the receipts map.
	receipt, exists := receipts[id]
	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	//Calculate points based on established rules.
	points := calculatePoints(receipt)

	//Spin up a response body in JSON.
	response := PointsResponse{Points: points}

	//Send the response.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Function to calculate the points given a receipt.
func calculatePoints(receipt *Receipt) int {
	//Regular expression to trim non-alphanumeric characters from retailer string.
	var nonAlphanumericRegex = regexp.MustCompile(`[^\p{L}\p{N} ]+`)

	//Trim all non-alphanumeric characters from retailer string and trim all whitespace.
	var length = strings.TrimSpace(nonAlphanumericRegex.ReplaceAllString(receipt.Retailer, ""))
	length = strings.Replace(length, " ", "", -1)

	//Calculate points based on length of trimmed retailer name string.
	var points = len(length)

	//If the total purchase amount is an even dollar ammount, add 50 points.
	if receipt.Total == math.Trunc(receipt.Total) {
		points += 50
	}

	//If the total purchase amount is a factor of 0.25, add 25 points.
	if math.Mod(receipt.Total, 0.25) == 0 {
		points += 25
	}

	//5 points for every two items on the receipt.
	points += (len(receipt.Items) / 2) * 5

	//If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer
	// The result is the number of points earned.
	for i := 0; i < len(receipt.Items); i++ {
		if len(strings.TrimSpace(receipt.Items[i].Description))%3 == 0 {
			points += int(math.Ceil(receipt.Items[i].Price * 0.2))
		}
	}

	//Date format.
	format := "2006-01-02"

	after, err := time.Parse("15:04", "14:00")
	before, err := time.Parse("15:04", "16:00")

	purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime)
	purchaseDate, err := time.Parse(format, receipt.PurchaseDate)
	fmt.Println(purchaseDate)

	if err != nil {
		fmt.Println(err)
	}

	//6 points if the day in the purchase date is odd.
	if purchaseDate.Day()%2 != 0 {
		points += 6
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	if purchaseTime.After(after) && purchaseTime.Before(before) {

		points += 10
	}

	//return the calculated points
	return points
}

func main() {

	//Implement a new HTTP request router r.
	r := mux.NewRouter()

	//Handle any new receipt (POST) request given as a JSON.
	r.HandleFunc("/receipts/process", processReceiptsHandler).Methods("POST")

	//Handle any new points (GET) request given a valid receipt id.
	r.HandleFunc("/receipts/{id}/points", getPointsHandler).Methods("GET")

	http.Handle("/", r)

	//Listen and service any request on port 3000.
	fmt.Println("Server listening on port 3000...")
	http.ListenAndServe(":3000", nil)
}

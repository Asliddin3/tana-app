package equipments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type GetOrderReq struct {
	apiHost string `json:"-"`
	Name    string `json:"name" form:"name"`
	Barcode string `json:"barcode" form:"barcode"`
}
type AnalizatorOrderRes struct {
	Gender      string                   `json:"gender"`
	Biomaterial string                   `json:"biomaterial"`
	IsPregnant  bool                     `json:"isPregnant"`
	Tests       []AnalizatorOrderTestRes `json:"tests"`
}

type AnalizatorOrderTestRes struct {
	Code string `json:"code"`
}

func getOrder(apiHost string, name string, barcode string) (AnalizatorOrderRes, error) {
	baseURL := fmt.Sprintf("http://%s/api/v1/analizator-result/order", apiHost)

	// Create a map for query parameters
	queryParams := map[string]string{
		"name":    name,
		"barcode": barcode,
	}

	// Encode the query parameters
	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}

	// Append the encoded query parameters to the base URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())

	// Send GET request
	response, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
		return AnalizatorOrderRes{}, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return AnalizatorOrderRes{}, err
	}
	res := AnalizatorOrderRes{}
	err = json.Unmarshal(body, &res)
	// Print the response
	fmt.Println("Response Body:", string(body))
	return res, nil
}

func saveOrder(data AnalizatorResultReq) error {
	url := fmt.Sprintf("http://%s/api/v1/analizator-result/order", data.apiHost)

	// Create an instance of YourStruct

	// Marshal the struct into JSON
	requestBody, err := json.Marshal(&data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	// Send POST request
	response, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error sending POST request:", err)
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		return nil
	}
	var bodyData []byte
	response.Body.Read(bodyData)
	return fmt.Errorf("got unexpected status code %d %s", response.StatusCode, string(bodyData))
	// // Read the response body
	// body, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	fmt.Println("Error reading response body:", err)
	// 	return err
	// }

}

type AnalizatorResultReq struct {
	apiHost string    `json:"-"`
	Barcode string    `json:"barcode" form:"barcode"`
	Name    string    `json:"name" form:"name"`
	Items   []Results `json:"items" form:"items"`
}
type AnalizatorResultItemReq struct {
	Code string `json:"code" form:"code"`
	Unit string `json:"unit" form:"unit"`
	Res  string `json:"res" form:"res"`
	Norm string `json:"norm" form:"norm"`
	Flag string `json:"flag" form:"flag"`
}

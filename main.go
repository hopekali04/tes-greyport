package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type GreypotClient interface {
	GeneratePDF(templateId, templateContent string, data interface{}) (*ExportResponse, error)
}

type GreypotHttpClient struct {
	BaseApiURL   string
	ObjectMapper *json.Decoder
}

func NewGreypotHttpClient(apiURL string, objectMapper *json.Decoder) *GreypotHttpClient {
	return &GreypotHttpClient{
		BaseApiURL:   apiURL,
		ObjectMapper: objectMapper,
	}
}

func (c *GreypotHttpClient) GeneratePDF(templateId, templateContent string, data interface{}) (*ExportResponse, error) {
	request := GeneratePDFRequest{
		Name:     templateId,
		Template: templateContent,
		Data:     data,
	}

	url := fmt.Sprintf("%s/_studio/generate/pdf/%s", c.BaseApiURL, request.Name)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Sending request body %s\n", requestBody)

	httpClient := http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response ExportResponse
	// err = c.ObjectMapper.Decode(respBody, &response)
	// fmt.Println("Error sending request body", string(respBody))
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		fmt.Println("Error unmarshalling response")
		return nil, err
	}

	return &response, nil
}

type ExportResponse struct {
	ReportID string `json:"reportId"`
	Type     string `json:"type"`
	Data     string `json:"data"`
}

func (r *ExportResponse) DataAsByteArray() ([]byte, error) {
	return base64.StdEncoding.DecodeString(r.Data)
}

type GeneratePDFRequest struct {
	Name     string      `json:"Name"`
	Template string      `json:"Template"`
	Data     interface{} `json:"Data"`
}

func NewGeneratePDFRequest(name, template string, data interface{}) *GeneratePDFRequest {
	return &GeneratePDFRequest{
		Name:     name,
		Template: template,
		Data:     data,
	}
}

func main() {
	// Initialize GreypotHttpClient
	apiURL := "https://greypot-studio.fly.dev"
	httpClient := NewGreypotHttpClient(apiURL, json.NewDecoder(os.Stdin))

	// Sample data for PDF generation
	templateID := "test.html"
	// Read HTML content from file
	htmlContent, err := os.ReadFile("sample.html")
	if err != nil {
		fmt.Println("Error reading HTML file:", err)
		return
	}
	
	data := map[string]interface{}{
		"invoiceId": "SH200992",
		"lineItems": []map[string]interface{}{
			{
				"name":       "Large box of gold",
				"netWeight":  "0.1 KG",
				"quantity":   3,
				"sku":        "A00005454",
				"totalValue": "30$",
				"unitValue":  "10$",
			},
			{
				"name":       "Fresh Air",
				"netWeight":  "0.3 KG",
				"quantity":   10,
				"sku":        "A0000522354",
				"totalValue": "150$",
				"unitValue":  "15$",
			},
		},
	}

	// Generate PDF
	response, err := httpClient.GeneratePDF(templateID, string(htmlContent), data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Report ID:", response.ReportID)
	fmt.Println("Type:", response.Type)

	// saving generated pdf
	byteData, err := response.DataAsByteArray()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	err = os.WriteFile("generated.pdf", byteData, 0644)
	if err != nil {
		fmt.Println("Error saving PDF:", err)
		return
	}

	fmt.Println("PDF saved as generated.pdf")
}

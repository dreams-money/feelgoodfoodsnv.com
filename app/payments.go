package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Payment struct {
	Status string `json:"status"`
}

type PaymentResponse struct {
	Payment Payment `json:"payment"`
}

func SubmitPayment(payToken, amount string) (PaymentResponse, error) {
	var response PaymentResponse

	req, err := makePaymentRequest(payToken, amount)
	if err != nil {
		return response, err
	}

	client := &http.Client{}
	paymentResponse, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer paymentResponse.Body.Close()

	paymentResponseBody, err := ioutil.ReadAll(paymentResponse.Body)
	if err != nil {
		return response, err
	}

	json.Unmarshal(paymentResponseBody, &response)

	return response, nil
}

func makePaymentRequest(payToken, amount string) (*http.Request, error) {
	var postUrl, privatePaymentToken string
	if strings.ToLower(serverConfiguration.Environment) == "dev" {
		postUrl = "https://connect.squareupsandbox.com/v2/payments"
		privatePaymentToken = serverConfiguration.PrivatePaymentKeys.Development
	} else {
		postUrl = "https://connect.squareup.com/v2/payments"
		privatePaymentToken = serverConfiguration.PrivatePaymentKeys.Production
	}

	amount = strings.Replace(amount, ".", "", 1)

	body := []byte(fmt.Sprintf(`{
		"amount_money": {
			"amount": %v,
			"currency": "USD"
		},
		"idempotency_key": "%v",
		"source_id": "%v"
		}`, amount, uuid.New(), payToken))

	req, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Square-Version", "2023-03-15")
	req.Header.Add("Authorization", "Bearer "+privatePaymentToken)

	return req, nil
}

package payment

import (
	"fmt"
	"strings"

	"saweria-be/internal/domain"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

// ChargeResult is defined in domain.ChargeResult — see internal/domain/types.go.

type midtransClient struct {
	coreClient coreapi.Client
}

func NewMidtransClient(serverKey, clientKey, env string) *midtransClient {
	var mtEnv midtrans.EnvironmentType
	if env == "production" {
		mtEnv = midtrans.Production
	} else {
		mtEnv = midtrans.Sandbox
	}
	var client coreapi.Client
	client.New(serverKey, mtEnv)
	return &midtransClient{coreClient: client}
}

// CreateCharge submits a Core API charge.
// paymentType: "bank_transfer" | "echannel" | "gopay" | "shopeepay"
// bank: "bca" | "bni" | "bri" | "permata" | "mandiri"  (only for bank_transfer / echannel)
func (m *midtransClient) CreateCharge(orderID string, amount int64, donorName, paymentType, bank string) (*domain.ChargeResult, error) {
	txDetail := midtrans.TransactionDetails{
		OrderID:  orderID,
		GrossAmt: amount,
	}
	customer := &midtrans.CustomerDetails{FName: donorName}

	var req *coreapi.ChargeReq

	switch strings.ToLower(paymentType) {
	case "bank_transfer":
		req = &coreapi.ChargeReq{
			PaymentType:        coreapi.PaymentTypeBankTransfer,
			TransactionDetails: txDetail,
			CustomerDetails:    customer,
			BankTransfer: &coreapi.BankTransferDetails{
				Bank: midtrans.Bank(strings.ToLower(bank)),
			},
		}
	case "echannel": // Mandiri Bill Payment
		req = &coreapi.ChargeReq{
			PaymentType:        coreapi.PaymentTypeEChannel,
			TransactionDetails: txDetail,
			CustomerDetails:    customer,
			EChannel: &coreapi.EChannelDetail{
				BillInfo1: "Donation",
				BillInfo2: donorName,
			},
		}
	case "gopay":
		req = &coreapi.ChargeReq{
			PaymentType:        coreapi.PaymentTypeGopay,
			TransactionDetails: txDetail,
			CustomerDetails:    customer,
			Gopay:              &coreapi.GopayDetails{EnableCallback: false},
		}
	case "shopeepay":
		req = &coreapi.ChargeReq{
			PaymentType:        coreapi.PaymentTypeShopeepay,
			TransactionDetails: txDetail,
			CustomerDetails:    customer,
		}
	case "qris":
		req = &coreapi.ChargeReq{
			PaymentType:        coreapi.PaymentTypeQris,
			TransactionDetails: txDetail,
			CustomerDetails:    customer,
		}
	default:
		return nil, fmt.Errorf("midtrans.CreateCharge: unsupported payment type %q", paymentType)
	}

	resp, err := m.coreClient.ChargeTransaction(req)
	if err != nil {
		return nil, fmt.Errorf("midtrans.CreateCharge: %w", err)
	}

	return parseChargeResponse(resp), nil
}

func parseChargeResponse(resp *coreapi.ChargeResponse) *domain.ChargeResult {
	result := &domain.ChargeResult{TransactionID: resp.TransactionID}

	// Bank transfer — resp.VaNumbers is populated
	if len(resp.VaNumbers) > 0 {
		result.Bank = resp.VaNumbers[0].Bank
		result.VANumber = resp.VaNumbers[0].VANumber
	}

	// Mandiri echannel
	if resp.BillerCode != "" {
		result.BillerCode = resp.BillerCode
		result.BillKey = resp.BillKey
	}

	// E-wallet actions
	for _, action := range resp.Actions {
		switch action.Name {
		case "generate-qr-code":
			result.QRCodeURL = action.URL
		case "deeplink-redirect":
			result.DeepLinkURL = action.URL
		}
	}

	return result
}

package payments

import (
	"fa-middleware/config"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
)

// func MakePaymentIntent(conf config.Config, userId string) (pi *stripe.PaymentIntent, err error) {
// 	// Set your secret key. Remember to switch to your live secret key in production.
// 	// See your keys here: https://dashboard.stripe.com/account/apikeys
// 	stripe.Key = conf.StripeSecretKey

// 	params := &stripe.PaymentIntentParams{
// 		Amount:   stripe.Int64(1000),
// 		Currency: stripe.String(string(stripe.CurrencyUSD)),
// 		PaymentMethodTypes: stripe.StringSlice([]string{
// 			"card",
// 		}),
// 		ReceiptEmail: stripe.String("jenny.rosen@example.com"),
// 	}

// 	pi, err = paymentintent.New(params)
// 	if err != nil {
// 		return pi, fmt.Errorf("failed to make paymentintent: %v", err.Error())
// 	}

// 	return pi, nil
// }

type CreateCheckoutSessionResponse struct {
	SessionID string `json:"id"`
}

func CreateCheckoutSession(c *gin.Context, conf config.Config) error {
	// domain := "http://localhost:4242"
	sc := &client.API{}
	sc.Init(conf.StripeSecretKey, nil)
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("T-shirt"),
					},
					UnitAmount: stripe.Int64(2000),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(conf.FullDomainURL + "/pages/t/stripesuccess.html"),
		CancelURL:  stripe.String(conf.FullDomainURL + "/pages/t/stripecancel.html"),
	}

	// session, err := session.New(params)
	session, err := sc.CheckoutSessions.New(params)
	if err != nil {
		return fmt.Errorf("session.New: %v", err.Error())
	}

	data := CreateCheckoutSessionResponse{
		SessionID: session.ID,
	}

	c.JSON(200, data)
	return nil
}

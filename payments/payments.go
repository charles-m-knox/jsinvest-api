package payments

import (
	"fa-middleware/config"

	"fmt"

	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/v72"
)

func MakePaymentIntent(conf config.Config, userId string) (pi *stripe.PaymentIntent, err error) {
	// Set your secret key. Remember to switch to your live secret key in production.
	// See your keys here: https://dashboard.stripe.com/account/apikeys
	stripe.Key = ""

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(1000),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		ReceiptEmail: stripe.String("jenny.rosen@example.com"),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return pi, fmt.Errorf("failed to make paymentintent: %v", err.Error())
	}

	return pi, nil
}

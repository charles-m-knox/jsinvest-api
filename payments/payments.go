package payments

import (
	"fa-middleware/config"
	"fa-middleware/models"
	"log"

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

// https://stripe.com/docs/api/products/retrieve
func GetProducts(conf config.Config) (products []models.ProductSummary, err error) {
	sc := &client.API{}
	sc.Init(conf.StripeSecretKey, nil)
	params := &stripe.ProductParams{}
	for _, stripeProduct := range conf.StripeProducts {
		product, err := sc.Products.Get(stripeProduct.ProductID, params)
		if err != nil {
			return products, fmt.Errorf(
				"failed to get product from stripe by id %v: %v",
				stripeProduct.ProductID,
				err.Error(),
			)
		}
		// validate that the metadata for the product matches
		productAppID, ok := product.Metadata["appId"]
		if !ok || productAppID != conf.FusionAuthAppID {
			log.Printf(
				"appId=%v from stripe not defined or mismatched from configured app id=%v for product id %v",
				productAppID,
				conf.FusionAuthAppID,
				stripeProduct.ProductID,
			)
			continue
		}
		productTenantID, ok := product.Metadata["tenantId"]
		if !ok || productTenantID != conf.FusionAuthTenantID {
			log.Printf(
				"tenantId=%v from stripe not defined or mismatched from configured tenant id=%v for product id %v",
				productTenantID,
				conf.FusionAuthTenantID,
				stripeProduct.ProductID,
			)
			continue
		}
		if !product.Active {
			continue
		}
		imageURL := ""
		if len(product.Images) > 0 {
			imageURL = product.Images[0]
		}

		// get all the prices now
		productPrices := []models.ProductPrice{}
		for _, priceID := range stripeProduct.PriceIDs {
			stripePrice, err := sc.Prices.Get(priceID, &stripe.PriceParams{})
			if err != nil {
				log.Printf(
					"failed to get price id %v for product id %v: %v",
					priceID,
					stripeProduct.ProductID,
					err.Error(),
				)
			}
			if !stripePrice.Active {
				continue
			}
			if stripePrice.Product == nil {
				log.Printf(
					"stripe price %v doesn't have corresponding product, skipping",
					stripePrice.ID,
				)
			}
			if stripePrice.Product.ID != stripeProduct.ProductID {
				log.Printf(
					"price id %v and product id %v mismatch, ignoring",
					stripePrice.ID,
					stripeProduct.ProductID,
				)
			}
			// the price has been validated; now add it to the list of prices
			recurringInterval := ""
			recurringIntervalCount := int64(0)
			if stripePrice.Recurring != nil {
				recurringInterval = string(stripePrice.Recurring.Interval)
				recurringIntervalCount = stripePrice.Recurring.IntervalCount
			}
			productPrices = append(productPrices, models.ProductPrice{
				ID:                     stripePrice.ID,
				ProductID:              stripeProduct.ProductID,
				RecurringInterval:      recurringInterval,
				RecurringIntervalCount: recurringIntervalCount,
				Price:                  stripePrice.UnitAmount,
				PriceDecimal:           stripePrice.UnitAmountDecimal,
				Currency:               string(stripePrice.Currency),
				Description:            stripePrice.Nickname,
			})
		}
		products = append(products, models.ProductSummary{
			ID:          stripeProduct.ProductID,
			Name:        product.Name,
			Description: product.Description,
			ImageURL:    imageURL,
			Prices:      productPrices,
		})
	}

	return products, nil
}

type CreateCheckoutSessionResponse struct {
	SessionID string `json:"id"`
}

func CreateCheckoutSession(c *gin.Context, conf config.Config) error {
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

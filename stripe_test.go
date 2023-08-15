package stripe

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var s *Stripe

func TestMain(m *testing.M) {
	err := godotenv.Load("./../.env")
	if err != nil {
		fmt.Println("Error loading .env file")
		panic(err)
	}

	s = NewStripe(os.Getenv("STRIPE_PRIVATE_KEY"))

	code := m.Run()
	os.Exit(code)
}

// func TestStripe_GetAccountBalance(t *testing.T) {
// 	balance, err := s.GetAccountBalance()
// 	if err != nil {
// 		t.Errorf("Error getting account balance: %v", err)
// 	}

// 	fmt.Printf("Balance: %v\n", balance.Available[0].Amount)
// }

// // func TestStripe_CreatePaymentIntentToOurStripeAccount(t *testing.T) {

// // 	params := &stripe.PaymentIntentParams{
// // 		Amount:   stripe.Int64(1000),
// // 		Currency: stripe.String(string(stripe.CurrencyUSD)),
// // 	}

// // 	intent, err := s.CreatePaymentIntent(params)
// // 	if err != nil {
// // 		t.Errorf("Error creating payment intent: %v", err)
// // 	}

// // 	fmt.Printf("Intent: %v\n", intent)
// // }

// func TestStripe_ListCustomerPaymentMethods(t *testing.T) {
// 	params := &stripe.PaymentMethodListParams{
// 		Customer: stripe.String("cus_OREMvC15b9jfb6"),
// 		Type:     stripe.String(string(stripe.PaymentMethodTypeCard)),
// 	}

// 	i := s.ListPaymentMethods(params)
// 	for i.Next() {
// 		paymentMethod := i.PaymentMethod()
// 		fmt.Printf("Payment method: %v\n", paymentMethod)
// 	}
// }

// func TestStripe_CreateAndDeleteCustomer(t *testing.T) {
// 	params := &stripe.CustomerParams{
// 		Name:  stripe.String("Olly Filmer"),
// 		Email: stripe.String("olly.filmer@outlook.com"),
// 	}

// 	customer, err := s.CreateCustomer(params)
// 	if err != nil {
// 		t.Errorf("Error creating customer: %v", err)
// 	}

// 	fmt.Printf("Customer: %v\n", customer)

// 	customer, err = s.DeleteCustomer(customer.ID)
// 	if err != nil {
// 		t.Errorf("Error deleting customer: %v", err)
// 	}
// }

// func TestStripe_GeneratePriceObjectAndPaymentLinkThenDelete(t *testing.T) {

// 	priceParams := &stripe.PriceParams{
// 		UnitAmount: stripe.Int64(1000),
// 		Currency:   stripe.String(string(stripe.CurrencyUSD)),
// 		ProductData: &stripe.PriceProductDataParams{
// 			Name: stripe.String("Test Product"),
// 		},
// 	}

// 	price, err := s.CreatePrice(string(stripe.CurrencyUSD), priceParams)
// 	if err != nil {
// 		t.Errorf("Error creating price: %v", err)
// 	}

// 	items := []*stripe.PaymentLinkLineItemParams{
// 		{
// 			Price:    &price.ID,
// 			Quantity: stripe.Int64(1),
// 		},
// 	}

// 	link, err := s.CreatePaymentLink(
// 		items,
// 		&stripe.PaymentLinkParams{
// 		},
// 	)
// 	if err != nil {
// 		t.Errorf("Error creating payment link: %v", err)
// 	}

// 	fmt.Printf("Link: %v\n", link)
// 	fmt.Printf("Link URL: %v\n", link.URL)
// }

// func TestStripe_CreateCustomerPaymentMethod(t *testing.T) {
// 	params := &stripe.PaymentMethodParams{
// 		Type: stripe.String(string(stripe.PaymentMethodTypeCard)),
// 		Card: &stripe.PaymentMethodCardParams{
// 			CVC:      stripe.String("123"),
// 			ExpMonth: stripe.Int64(12),
// 			ExpYear:  stripe.Int64(2025),
// 			Number:   stripe.String("4242424242424242"),
// 		},
// 	}

// 	paymentMethod, err := s.CreatePaymentMethod(params)
// 	if err != nil {
// 		t.Errorf("Error creating customer payment method: %v", err)
// 	}

// 	fmt.Printf("Payment method: %v\n", paymentMethod)
// }

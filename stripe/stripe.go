package stripe

//TODO - Prepend all functions with Category and Section Names

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
	"github.com/stripe/stripe-go/v74/applicationfee"
	"github.com/stripe/stripe-go/v74/apps/secret"
	"github.com/stripe/stripe-go/v74/balancetransaction"
	"github.com/stripe/stripe-go/v74/bankaccount"
	"github.com/stripe/stripe-go/v74/capability"
	"github.com/stripe/stripe-go/v74/card"
	"github.com/stripe/stripe-go/v74/charge"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/client"
	"github.com/stripe/stripe-go/v74/countryspec"
	"github.com/stripe/stripe-go/v74/coupon"
	"github.com/stripe/stripe-go/v74/creditnote"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/customerbalancetransaction"
	"github.com/stripe/stripe-go/v74/customercashbalancetransaction"
	"github.com/stripe/stripe-go/v74/dispute"
	"github.com/stripe/stripe-go/v74/event"
	"github.com/stripe/stripe-go/v74/feerefund"
	"github.com/stripe/stripe-go/v74/file"
	"github.com/stripe/stripe-go/v74/filelink"
	"github.com/stripe/stripe-go/v74/invoice"
	"github.com/stripe/stripe-go/v74/invoiceitem"
	"github.com/stripe/stripe-go/v74/issuing/authorization"
	"github.com/stripe/stripe-go/v74/issuing/cardholder"
	"github.com/stripe/stripe-go/v74/issuing/transaction"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/paymentlink"
	"github.com/stripe/stripe-go/v74/paymentmethod"
	"github.com/stripe/stripe-go/v74/payout"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/promotioncode"
	"github.com/stripe/stripe-go/v74/quote"
	"github.com/stripe/stripe-go/v74/radar/earlyfraudwarning"
	"github.com/stripe/stripe-go/v74/radar/valuelist"
	"github.com/stripe/stripe-go/v74/radar/valuelistitem"
	"github.com/stripe/stripe-go/v74/refund"
	"github.com/stripe/stripe-go/v74/review"
	"github.com/stripe/stripe-go/v74/setupattempt"
	"github.com/stripe/stripe-go/v74/setupintent"
	"github.com/stripe/stripe-go/v74/shippingrate"
	"github.com/stripe/stripe-go/v74/subscription"
	"github.com/stripe/stripe-go/v74/subscriptionitem"
	"github.com/stripe/stripe-go/v74/subscriptionschedule"
	"github.com/stripe/stripe-go/v74/taxcode"
	"github.com/stripe/stripe-go/v74/taxid"
	"github.com/stripe/stripe-go/v74/taxrate"
	"github.com/stripe/stripe-go/v74/terminal/configuration"
	"github.com/stripe/stripe-go/v74/terminal/location"
	"github.com/stripe/stripe-go/v74/terminal/reader"
	"github.com/stripe/stripe-go/v74/testhelpers/testclock"
	"github.com/stripe/stripe-go/v74/topup"
	"github.com/stripe/stripe-go/v74/transfer"
	"github.com/stripe/stripe-go/v74/transferreversal"
	"github.com/stripe/stripe-go/v74/usagerecordsummary"

	bpc "github.com/stripe/stripe-go/v74/billingportal/configuration"
	issuingCard "github.com/stripe/stripe-go/v74/issuing/card"
	issuingDispute "github.com/stripe/stripe-go/v74/issuing/dispute"
)

type Stripe struct {
	client *client.API
}

func NewStripe(secretKey string) *Stripe {
	s := &Stripe{}
	sc := &client.API{}
	sc.Init(secretKey, nil)
	s.client = sc

	return s
}

func (s *Stripe) handleStripeError(err error) error {
	if err != nil {
		// Try to safely cast a generic error to a stripe.Error so that we can get at
		// some additional Stripe-specific information about what went wrong.
		if stripeErr, ok := err.(*stripe.Error); ok {
			// The Code field will contain a basic identifier for the failure.
			switch stripeErr.Code {
			case stripe.ErrorCodeCardDeclined:
				color.Red("Card was declined")
			case stripe.ErrorCodeExpiredCard:
				color.Red("Card has expired")
			case stripe.ErrorCodeIncorrectCVC:
				color.Red("Incorrect CVC")
			case stripe.ErrorCodeIncorrectZip:
				color.Red("Incorrect ZIP")
			case stripe.ErrorCodeAmountTooLarge:
				color.Red("Amount too large")
			case stripe.ErrorCodeAmountTooSmall:
				color.Red("Amount too small")
			case stripe.ErrorCodeBalanceInsufficient:
				color.Red("Insufficient balance")
			case stripe.ErrorCodeMissing:
				color.Red("Missing")
			case stripe.ErrorCodeProcessingError:
				color.Red("Processing error")
			case stripe.ErrorCodeRateLimit:
				color.Red("Rate limit")
			case stripe.ErrorCodeAPIKeyExpired:
				color.Red("API key expired")
			//more...
			default:
				color.Red("Other error occurred...")
				color.Red(stripeErr.Error())
			}

			color.Cyan("Request ID : %v", stripeErr.RequestID)

			// The Err field can be coerced to a more specific error type with a type
			// assertion. This technique can be used to get more specialized
			// information for certain errors.
			if cardErr, ok := stripeErr.Err.(*stripe.CardError); ok {
				fmt.Printf("Card was declined with code: %v\n", cardErr.DeclineCode)
			} else {
				fmt.Printf("Other Stripe error occurred: %v\n", stripeErr.Error())
			}
		} else {
			fmt.Printf("Other error occurred: %v\n", err.Error())
		}

	}

	return nil
}

func (s *Stripe) genIdempotency() string {
	return stripe.NewIdempotencyKey()
}

//SECTION - BALANCE
/*
	This is an object representing your Stripe balance. You can retrieve it to see the balance currently on your Stripe account.

	You can also retrieve the balance history, which contains a list of transactions that contributed to the balance (charges, payouts, and so forth).

	The available and pending amounts for each currency are broken down further by payment source types.
*/

// https://stripe.com/docs/api/balance/balance_object?lang=go
func (s *Stripe) GetAccountBalance() (*stripe.Balance, error) {
	balance, err := s.client.Balance.Get(nil)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

//SECTION - BALANCE TRANSACTIONS
/*
	Balance transactions represent funds moving through your Stripe account. They're created for every type of transaction that comes into or flows out of your Stripe account balance.
*/

// https://stripe.com/docs/api/balance_transactions/retrieve?lang=go
func (s *Stripe) GetBalanceTransaction(id string) (*stripe.BalanceTransaction, error) {
	balanceTransaction, err := s.client.BalanceTransactions.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return balanceTransaction, nil
}

//https://stripe.com/docs/api/balance_transactions/list?lang=go
// Iterate through balance transactions
/*
	for i.Next() {
		bt := i.BalanceTransaction()
		fmt.Printf("Transaction: %v\n", bt.ID)
	}
*/
func (s *Stripe) ListBalanceTransactions() *balancetransaction.Iter {
	params := &stripe.BalanceTransactionListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.BalanceTransactions.List(params)
	return i
}

//SECTION - CHARGES
/*
	The Charge object represents a single attempt to move money into your Stripe account. PaymentIntent confirmation is the most common way to create Charges, but transferring money to a different Stripe account through Connect also creates Charges. Some legacy payment flows create Charges directly, which is not recommended for new integrations.
*/

// https://stripe.com/docs/api/charges/create?lang=go
func (s *Stripe) CreateCharge(params *stripe.ChargeParams) (*stripe.Charge, error) {
	c, err := s.client.Charges.New(params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/charges/retrieve?lang=go
func (s *Stripe) GetCharge(id string) (*stripe.Charge, error) {
	charge, err := s.client.Charges.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return charge, nil
}

// https://stripe.com/docs/api/charges/update?lang=go
func (s *Stripe) UpdateCharge(id string, params *stripe.ChargeParams) (*stripe.Charge, error) {
	c, err := s.client.Charges.Update(id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/charges/capture?lang=go
func (s *Stripe) CaptureCharge(id string, opt_params *stripe.ChargeCaptureParams) (*stripe.Charge, error) {
	c, err := s.client.Charges.Capture(id, opt_params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/charges/list?lang=go
// Iterate through charges
/*
	for i.Next() {
		c := i.Charge()
		fmt.Printf("Charge: %v\n", c.ID)
	}
*/
func (s *Stripe) ListCharges() *charge.Iter {
	params := &stripe.ChargeListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.Charges.List(params)
	return i
}

// https://stripe.com/docs/api/charges/search?lang=go
/*
for iter.Next() {
  result := iter.Current()
}
*/
func (s *Stripe) SearchCharges(query string) *charge.SearchIter {
	params := &stripe.ChargeSearchParams{}
	params.Query = *stripe.String(query)
	i := s.client.Charges.Search(params)
	return i
}

//SECTION - CUSTOMERS
/*
	This object represents a customer of your business. It lets you create recurring charges and track payments that belong to the same customer.
*/

// https://stripe.com/docs/api/customers/create
func (s *Stripe) CreateCustomer(params *stripe.CustomerParams) (*stripe.Customer, error) {
	c, err := s.client.Customers.New(params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/customers/retrieve
func (s *Stripe) GetCustomer(id string) (*stripe.Customer, error) {
	customer, err := s.client.Customers.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// https://stripe.com/docs/api/customers/update
func (s *Stripe) UpdateCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	c, err := s.client.Customers.Update(id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/customers/delete
func (s *Stripe) DeleteCustomer(id string) (*stripe.Customer, error) {
	c, err := s.client.Customers.Del(id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/customers/list
// Iterate through customers
func (s *Stripe) ListCustomers() *customer.Iter {
	params := &stripe.CustomerListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.Customers.List(params)
	return i
}

// https://stripe.com/docs/api/customers/search
func (s *Stripe) SearchCustomers(query string) *customer.SearchIter {
	params := &stripe.CustomerSearchParams{}
	params.Query = *stripe.String(query)
	i := s.client.Customers.Search(params)
	return i
}

//SECTION - DISPUTES
/*
	A dispute occurs when a customer questions your charge with their card issuer. When this happens, you're given the opportunity to respond to the dispute with evidence that shows that the charge is legitimate. You can find more information about the dispute process in our Disputes and Fraud documentation.
*/

// https://stripe.com/docs/api/disputes/retrieve
func (s *Stripe) GetDispute(id string) (*stripe.Dispute, error) {
	dispute, err := s.client.Disputes.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return dispute, nil
}

// https://stripe.com/docs/api/disputes/update
func (s *Stripe) UpdateDispute(id string, params *stripe.DisputeParams) (*stripe.Dispute, error) {
	d, err := s.client.Disputes.Update(id, params)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// https://stripe.com/docs/api/disputes/close
func (s *Stripe) CloseDispute(id string) (*stripe.Dispute, error) {
	d, err := s.client.Disputes.Close(id, nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// https://stripe.com/docs/api/disputes/list
// Iterate through disputes
/*
for i.Next() {
  d := i.Dispute()
}
*/
func (s *Stripe) ListDisputes() *dispute.Iter {
	params := &stripe.DisputeListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.Disputes.List(params)
	return i
}

//SECTION - EVENTS
/*
	Events are our way of letting you know when something interesting happens in your account. When an interesting event occurs, we create a new Event object. For example, when a charge succeeds, we create a charge.succeeded event; and when an invoice payment attempt fails, we create an invoice.payment_failed event. Note that many API requests may cause multiple events to be created. For example, if you create a new subscription for a customer, you will receive both a customer.subscription.created event and a charge.succeeded event.

	Events occur when the state of another API resource changes. The state of that resource at the time of the change is embedded in the event's data field. For example, a charge.succeeded event will contain a charge, and an invoice.payment_failed event will contain an invoice.

	As with other API resources, you can use endpoints to retrieve an individual event or a list of events from the API. We also have a separate webhooks system for sending the Event objects directly to an endpoint on your server. Webhooks are managed in your account settings, and our Using Webhooks guide will help you get set up.

	When using Connect, you can also receive notifications of events that occur in connected accounts. For these events, there will be an additional account attribute in the received Event object.

	NOTE: Right now, access to events through the Retrieve Event API is guaranteed only for 30 days.
*/

// https://stripe.com/docs/api/events/retrieve
func (s *Stripe) GetEvent(id string) (*stripe.Event, error) {
	event, err := s.client.Events.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *Stripe) HandleEvent(event *stripe.Event) error {
	eventType := event.Type

	switch eventType {
	case "charge.succeeded":
		fmt.Println("Charge succeeded")
	case "charge.failed":
		fmt.Println("Charge failed")
	//...
	default:
		fmt.Println("Unhandled event type:")
		fmt.Println(eventType)
	}

	return nil
}

//https://stripe.com/docs/api/events/list
// Iterate through events
/*
for i.Next() {
  e := i.Event()
}
*/
func (s *Stripe) ListEvents() *event.Iter {
	params := &stripe.EventListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.Events.List(params)
	return i
}

//SECTION - FILES
/*
	This is an object representing a file hosted on Stripe's servers. The file may have been uploaded by yourself using the create file request (for example, when uploading dispute evidence) or it may have been created by Stripe (for example, the results of a Sigma scheduled query).
*/

// https://stripe.com/docs/api/files/create
func (s *Stripe) CreateFile(pathToFile string, fileName string, purpose stripe.FilePurpose) (*stripe.File, error) {
	fp, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	params := &stripe.FileParams{}
	params.FileReader = fp
	params.Filename = stripe.String(fileName)
	params.Purpose = stripe.String(string(purpose))

	f, err := s.client.Files.New(params)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// https://stripe.com/docs/api/files/retrieve
func (s *Stripe) GetFile(id string) (*stripe.File, error) {
	f, err := s.client.Files.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// https://stripe.com/docs/api/files/list
// Iterate through files
/*
for i.Next() {
  f := i.File()
}
*/
func (s *Stripe) ListFiles() *file.Iter {
	params := &stripe.FileListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.Files.List(params)
	return i
}

//SECTION - FILE LINKS
/*
	To share the contents of a File object with non-Stripe users, you can create a FileLink. FileLinks contain a URL that can be used to retrieve the contents of the file without authentication.
*/

// https://stripe.com/docs/api/file_links/create
func (s *Stripe) CreateFileLink(fileId string) (*stripe.FileLink, error) {
	params := &stripe.FileLinkParams{}
	params.File = stripe.String(fileId)

	fl, err := s.client.FileLinks.New(params)
	if err != nil {
		return nil, err
	}

	return fl, nil
}

// https://stripe.com/docs/api/file_links/retrieve
func (s *Stripe) GetFileLink(id string) (*stripe.FileLink, error) {
	fl, err := s.client.FileLinks.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return fl, nil
}

// https://stripe.com/docs/api/file_links/update
func (s *Stripe) UpdateFileLink(id string, params *stripe.FileLinkParams) (*stripe.FileLink, error) {
	fl, err := s.client.FileLinks.Update(id, params)
	if err != nil {
		return nil, err
	}

	return fl, nil
}

// https://stripe.com/docs/api/file_links/list
// Iterate through file links
/*
for i.Next() {
  fl := i.FileLink()
}
*/
func (s *Stripe) ListFileLinks() *filelink.Iter {
	params := &stripe.FileLinkListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.FileLinks.List(params)
	return i
}

//SECTION - MANDATES
/*
	A Mandate is a record of the permission a customer has given you to debit their payment method.
*/

// https://stripe.com/docs/api/mandates/retrieve
func (s *Stripe) GetMandate(id string) (*stripe.Mandate, error) {
	mandate, err := s.client.Mandates.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return mandate, nil
}

//SECTION - PAYMENT INTENTS
/*
	A PaymentIntent guides you through the process of collecting a payment from your customer. We recommend that you create exactly one PaymentIntent for each order or customer session in your system. You can reference the PaymentIntent later to see the history of payment attempts for a particular session.

	A PaymentIntent transitions through multiple statuses throughout its lifetime as it interfaces with Stripe.js to perform authentication flows and ultimately creates at most one successful charge.
*/

// https://stripe.com/docs/api/payment_intents/create
/*
	//REQUIRED
	//- amount
	//- currency - 3-letter ISO code for currency
*/
// https://stripe.com/docs/currencies
func (s *Stripe) CreatePaymentIntent(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.New(params)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// https://stripe.com/docs/api/payment_intents/retrieve
func (s *Stripe) GetPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// https://stripe.com/docs/api/payment_intents/update
func (s *Stripe) UpdatePaymentIntent(id string, params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.Update(id, params)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// https://stripe.com/docs/api/payment_intents/confirm
// https://stripe.com/docs/payments/payment-intents/creating-payment-intents#creating-for-automatic
func (s *Stripe) ConfirmPaymentIntent(id string, params *stripe.PaymentIntentConfirmParams) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.Confirm(id, params)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// https://stripe.com/docs/api/payment_intents/capture
func (s *Stripe) CapturePaymentIntent(id string, opt_params *stripe.PaymentIntentCaptureParams) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.Capture(id, opt_params)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// https://stripe.com/docs/api/payment_intents/cancel
func (s *Stripe) CancelPaymentIntent(id string, params *stripe.PaymentIntentCancelParams) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.Cancel(id, params)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// https://stripe.com/docs/api/payment_intents/list
// Iterate through payment intents
/*
for i.Next() {
	  pi := i.PaymentIntent()
}
*/
func (s *Stripe) ListPaymentIntents() *paymentintent.Iter {
	params := &stripe.PaymentIntentListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.PaymentIntents.List(params)
	return i
}

// TERMINAL ONLY
// https://stripe.com/docs/api/payment_intents/increment_authorization
func (s *Stripe) IncrementPaymentIntent(id string, params *stripe.PaymentIntentIncrementAuthorizationParams) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.IncrementAuthorization(id, params)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// https://stripe.com/docs/api/payment_intents/search
// Iterate through payment intents
/*
for iter.Next() {
  result := iter.Current()
}
*/
func (s *Stripe) SearchPaymentIntents(query string) *paymentintent.SearchIter {
	params := &stripe.PaymentIntentSearchParams{}
	params.Query = *stripe.String(query)
	i := s.client.PaymentIntents.Search(params)
	return i
}

// https://stripe.com/docs/api/payment_intents/verify_microdeposits
func (s *Stripe) VerifyMicrodepositsOnPaymentIntent(id string, params *stripe.PaymentIntentVerifyMicrodepositsParams) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.VerifyMicrodeposits(id, params)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

// https://stripe.com/docs/api/payment_intents/apply_customer_balance
func (s *Stripe) ReconcileACustomerBalancePaymentIntent(id string, params *stripe.PaymentIntentApplyCustomerBalanceParams) (*stripe.PaymentIntent, error) {
	pi, err := s.client.PaymentIntents.ApplyCustomerBalance(id, params)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

//SECTION - SETUP INTENTS
/*
	A SetupIntent guides you through the process of setting up and saving a customer's payment credentials for future payments. For example, you could use a SetupIntent to set up and save your customer's card without immediately collecting a payment. Later, you can use PaymentIntents to drive the payment flow.

	Create a SetupIntent as soon as you're ready to collect your customer's payment credentials. Do not maintain long-lived, unconfirmed SetupIntents as they may no longer be valid. The SetupIntent then transitions through multiple statuses as it guides you through the setup process.

	Successful SetupIntents result in payment credentials that are optimized for future payments. For example, cardholders in certain regions may need to be run through Strong Customer Authentication at the time of payment method collection in order to streamline later off-session payments. If the SetupIntent is used with a Customer, upon success, it will automatically attach the resulting payment method to that Customer. We recommend using SetupIntents or setup_future_usage on PaymentIntents to save payment methods in order to prevent saving invalid or unoptimized payment methods.

	By using SetupIntents, you ensure that your customers experience the minimum set of required friction, even as regulations change over time.
*/

// https://stripe.com/docs/api/setup_intents/create
func (s *Stripe) CreateSetupIntent(params *stripe.SetupIntentParams) (*stripe.SetupIntent, error) {
	si, err := s.client.SetupIntents.New(params)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/setup_intents/retrieve
func (s *Stripe) GetSetupIntent(id string) (*stripe.SetupIntent, error) {
	si, err := s.client.SetupIntents.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/setup_intents/update
func (s *Stripe) UpdateSetupIntent(id string, params *stripe.SetupIntentParams) (*stripe.SetupIntent, error) {
	si, err := s.client.SetupIntents.Update(id, params)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/setup_intents/confirm
func (s *Stripe) ConfirmSetupIntent(id string, params *stripe.SetupIntentConfirmParams) (*stripe.SetupIntent, error) {
	si, err := s.client.SetupIntents.Confirm(id, params)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/setup_intents/confirm
func (s *Stripe) CancelSetupIntent(id string, params *stripe.SetupIntentCancelParams) (*stripe.SetupIntent, error) {
	si, err := s.client.SetupIntents.Cancel(id, params)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/setup_intents/list
// Iterate through setup intents
/*
for i.Next() {
	  si := i.SetupIntent()
}
*/
func (s *Stripe) ListSetupIntents() *setupintent.Iter {
	params := &stripe.SetupIntentListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.SetupIntents.List(params)
	return i
}

// https://stripe.com/docs/api/setup_intents/verify_microdeposits
func (s *Stripe) VerifyMicrodepositsOnSetupIntent(id string, opt_params *stripe.SetupIntentVerifyMicrodepositsParams) (*stripe.SetupIntent, error) {
	si, err := s.client.SetupIntents.VerifyMicrodeposits(id, opt_params)
	if err != nil {
		return nil, err
	}

	return si, nil
}

//SECTION - PAYMENT ATTEMPTS
/*
	A SetupAttempt describes one attempted confirmation of a SetupIntent, whether that confirmation was successful or unsuccessful. You can use SetupAttempts to inspect details of a specific attempt at setting up a payment method using a SetupIntent.
*/

// https://stripe.com/docs/api/setup_attempts/list
// Iterate through setup attempts
/*
for i.Next() {
	  sa := i.SetupAttempt()
}
*/
func (s *Stripe) ListSetupAttempts(id string) *setupattempt.Iter {
	params := &stripe.SetupAttemptListParams{}
	params.SetupIntent = stripe.String(id)

	i := s.client.SetupAttempts.List(params)
	return i
}

//SECTION - PAYOUTS
/*
	A Payout object is created when you receive funds from Stripe, or when you initiate a payout to either a bank account or debit card of a connected Stripe account. You can retrieve individual payouts, as well as list all payouts. Payouts are made on varying schedules, depending on your country and industry.
*/

// https://stripe.com/docs/api/payouts/create
func (s *Stripe) CreatePayout(params *stripe.PayoutParams) (*stripe.Payout, error) {
	p, err := s.client.Payouts.New(params)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/payouts/retrieve
func (s *Stripe) GetPayout(id string) (*stripe.Payout, error) {
	p, err := s.client.Payouts.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/payouts/update
func (s *Stripe) UpdatePayout(id string, params *stripe.PayoutParams) (*stripe.Payout, error) {
	p, err := s.client.Payouts.Update(id, params)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/payouts/list
// Iterate through payouts
/*
for i.Next() {
	  p := i.Payout()
}
*/
func (s *Stripe) ListPayouts() *payout.Iter {
	params := &stripe.PayoutListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.Payouts.List(params)
	return i
}

// https://stripe.com/docs/api/payouts/cancel
func (s *Stripe) CancelPayout(id string) (*stripe.Payout, error) {
	p, err := s.client.Payouts.Cancel(id, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/payouts/reverse
// CURRENTLY UNSUPPORTED IN GO

//SECTION - REFUNDS
/*
	Refund objects allow you to refund a charge that has previously been created but not yet refunded. Funds will be refunded to the credit or debit card that was originally charged.
*/

// https://stripe.com/docs/api/refunds/create
func (s *Stripe) CreateRefund(params *stripe.RefundParams) (*stripe.Refund, error) {
	r, err := s.client.Refunds.New(params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/refunds/retrieve
func (s *Stripe) GetRefund(id string) (*stripe.Refund, error) {
	r, err := s.client.Refunds.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/refunds/update
func (s *Stripe) UpdateRefund(id string, params *stripe.RefundParams) (*stripe.Refund, error) {
	r, err := s.client.Refunds.Update(id, params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/refunds/update
// CURENTLY UNSUPPORTED IN GO

// https://stripe.com/docs/api/refunds/list
// Iterate through refunds
/*
for i.Next() {
	  r := i.Refund()
}
*/
func (s *Stripe) ListRefunds() *refund.Iter {
	params := &stripe.RefundListParams{}
	params.Filters.AddFilter("limit", "", "3")
	params.Filters.AddFilter("include[]", "", "total_count")

	i := s.client.Refunds.List(params)
	return i
}

//SECTION - TOKENS
/*
	Tokenization is the process Stripe uses to collect sensitive card or bank account details, or personally identifiable information (PII), directly from your customers in a secure manner. A token representing this information is returned to your server to use. You should use our recommended payments integrations to perform this process client-side. This ensures that no sensitive card data touches your server, and allows your integration to operate in a PCI-compliant way.

	If you cannot use client-side tokenization, you can also create tokens using the API with either your publishable or secret API key. Keep in mind that if your integration uses this method, you are responsible for any PCI compliance that may be required, and you must keep your secret API key safe. Unlike with client-side tokenization, your customer's information is not sent directly to Stripe, so we cannot determine how it is handled or stored.

	Tokens cannot be stored or used more than once. To store card or bank account information for later use, you can create Customer objects or Custom accounts. Note that Radar, our integrated solution for automatic fraud protection, performs best with integrations that use client-side tokenization.
*/

// https://stripe.com/docs/api/tokens/create_card
func (s *Stripe) CreateCardToken(params *stripe.TokenParams) (*stripe.Token, error) {
	t, err := s.client.Tokens.New(params)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/tokens/create_bank_account
func (s *Stripe) CreateBankAccountToken(params *stripe.BankAccountParams) (*stripe.Token, error) {
	t, err := s.client.Tokens.New(&stripe.TokenParams{BankAccount: params})
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/tokens/create_pii
func (s *Stripe) CreatePiiToken(params *stripe.TokenPIIParams) (*stripe.Token, error) {
	t, err := s.client.Tokens.New(&stripe.TokenParams{PII: params})
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/tokens/create_account
// CURRENTLY UNSUPPORTED IN GO

// https://stripe.com/docs/api/tokens/create_person
// CURRENTLY UNSUPPORTED IN GO

// https://stripe.com/docs/api/tokens/create_cvc_update
func (s *Stripe) CreateCvcUpdateToken(params *stripe.TokenCVCUpdateParams) (*stripe.Token, error) {
	t, err := s.client.Tokens.New(&stripe.TokenParams{CVCUpdate: params})
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/tokens/retrieve
func (s *Stripe) GetToken(id string) (*stripe.Token, error) {
	t, err := s.client.Tokens.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

//SECTION - PAYMENT METHODS
/*
	PaymentMethod objects represent your customer's payment instruments. You can use them with PaymentIntents to collect payments or save them to Customer objects to store instrument details for future payments.
*/

// https://stripe.com/docs/api/payment_methods/create
/* required params:
- Type
*/
/*
   - acss_debit
   - affirm
   - afterpay_clearpay
   - alipay
   - au_becs_debit
   - bacs_debit
   - bancontact
   - blik
   - boleto
   - card
   - cashapp
   - customer_balance
   - eps
   - fpx
   - giropay
   - grabpay
   - ideal
   - klarna
   - konbini
   - link
   - oxxo
   - p24
   - paynow
   - paypal
   - pix
   - promptpay
   - sepa_debit
   - sofort
   - us_bank_account
   - wechat_pay
   - zip
*/
func (s *Stripe) CreatePaymentMethod(params *stripe.PaymentMethodParams) (*stripe.PaymentMethod, error) {
	pm, err := s.client.PaymentMethods.New(params)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

// https://stripe.com/docs/api/payment_methods/retrieve
func (s *Stripe) GetPaymentMethod(id string) (*stripe.PaymentMethod, error) {
	pm, err := s.client.PaymentMethods.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

// https://stripe.com/docs/api/payment_methods/customer
func (s *Stripe) GetCustomersPaymentMethod(pm_id, cus_id string) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodParams{
		Customer: stripe.String(cus_id),
	}

	pm, err := s.client.PaymentMethods.Get(pm_id, params)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

// https://stripe.com/docs/api/payment_methods/update
func (s *Stripe) UpdatePaymentMethod(id string, params *stripe.PaymentMethodParams) (*stripe.PaymentMethod, error) {
	pm, err := s.client.PaymentMethods.Update(id, params)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

// https://stripe.com/docs/api/payment_methods/list
// Iterate through payment methods
/*
for i.Next() {
	  pm := i.PaymentMethod()
}
*/
func (s *Stripe) ListPaymentMethods(params *stripe.PaymentMethodListParams) *paymentmethod.Iter {
	i := s.client.PaymentMethods.List(params)
	return i
}

// https://stripe.com/docs/api/payment_methods/customer_list
// Iterate through payment methods
/*
for i.Next() {
	  pm := i.PaymentMethod()
}
*/
func (s *Stripe) ListCustomersPaymentMethods(cus_id, paymentType string) *paymentmethod.Iter {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(cus_id),
		Type:     stripe.String(paymentType),
	}

	i := s.client.PaymentMethods.List(params)
	return i
}

// https://stripe.com/docs/api/payment_methods/attach
func (s *Stripe) AttachPaymentMethod(cus_id, pm_id string) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(cus_id),
	}

	pm, err := s.client.PaymentMethods.Attach(pm_id, params)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

// https://stripe.com/docs/api/payment_methods/detach
func (s *Stripe) DetachPaymentMethod(pm_id string) (*stripe.PaymentMethod, error) {
	pm, err := s.client.PaymentMethods.Detach(pm_id, nil)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

//SECTION - BANK ACCOUNTS
/*
	These bank accounts are payment methods on Customer objects.

	On the other hand External Accounts are transfer destinations on Account objects for Custom accounts. They can be bank accounts or debit cards as well, and are documented in the links above.
*/

// https://stripe.com/docs/api/customer_bank_accounts/create
// Tokens are prepended with btok_
func (s *Stripe) CreateCustomerBankAccountWithToken(cus_id, token string) (*stripe.BankAccount, error) {
	params := &stripe.BankAccountParams{
		Customer: stripe.String(cus_id),
		Token:    stripe.String(token),
	}

	ba, err := s.client.BankAccounts.New(params)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/customer_bank_accounts/create
func (s *Stripe) CreateCustomerBankAccount(params *stripe.BankAccountParams) (*stripe.BankAccount, error) {
	ba, err := s.client.BankAccounts.New(params)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/customer_bank_accounts/retrieve
// bank account id is prepended with ba_
func (s *Stripe) GetCustomerBankAccount(cus_id, ba_id string) (*stripe.BankAccount, error) {
	params := &stripe.BankAccountParams{
		Customer: stripe.String(cus_id),
	}

	ba, err := s.client.BankAccounts.Get(ba_id, params)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/customer_bank_accounts/update
func (s *Stripe) UpdateCustomerBankAccount(cus_id, ba_id string, params *stripe.BankAccountParams) (*stripe.BankAccount, error) {
	params.Customer = stripe.String(cus_id)

	ba, err := s.client.BankAccounts.Update(ba_id, params)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/customer_bank_accounts/verify
func (s *Stripe) VerifyCustomerBankAccount(cus_id, ba_id string, params *stripe.PaymentSourceVerifyParams) (*stripe.PaymentSource, error) {
	params.Customer = stripe.String(cus_id)

	ba, err := s.client.PaymentSources.Verify(ba_id, params)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/customer_bank_accounts/delete
func (s *Stripe) DeleteCustomerBankAccount(cus_id, ba_id string) (*stripe.BankAccount, error) {
	params := &stripe.BankAccountParams{
		Customer: stripe.String(cus_id),
	}

	ba, err := s.client.BankAccounts.Del(ba_id, params)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/customer_bank_accounts/list
// Iterate through bank accounts
/*
for i.Next() {
	  ba := i.BankAccount()
}
*/
func (s *Stripe) ListCustomerBankAccounts(cus_id string, params *stripe.BankAccountListParams) *bankaccount.Iter {
	params.Customer = stripe.String(cus_id)

	i := s.client.BankAccounts.List(params)
	return i
}

//SECTION - CASH BALANCE
/*
	A customer's Cash balance represents real funds. Customers can add funds to their cash balance by sending a bank transfer. These funds can be used for payment and can eventually be paid out to your bank account.
*/

// https://stripe.com/docs/api/cash_balance/retrieve
func (s *Stripe) GetCashBalance_preview(cus_id string) (*stripe.CashBalance, error) {
	params := &stripe.CashBalanceParams{
		Customer: stripe.String(cus_id),
	}

	cb, err := s.client.CashBalances.Get(params)
	if err != nil {
		return nil, err
	}

	return cb, nil
}

// https://stripe.com/docs/api/cash_balance/update
func (s *Stripe) UpdateCashBalanceSettings_preview(cus_id string, settings *stripe.CashBalanceSettingsParams) (*stripe.CashBalance, error) {
	params := &stripe.CashBalanceParams{
		Customer: stripe.String(cus_id),
		Settings: settings,
	}

	cb, err := s.client.CashBalances.Update(params)
	if err != nil {
		return nil, err
	}

	return cb, nil
}

//The Cash Balance Trasaction
/* Types
adjusted_for_overdraft
applied_to_payment
funded
funding_reversed
refunded_from_payment
return_cancelled
return_initiated
unapplied_from_payment
*/

// https://stripe.com/docs/api/cash_balance_transactions/retrieve
// cbt_id is prepended with ccsbtxn_
func (s *Stripe) GetCustomerCashBalanceTransaction(cus_id, cbt_id string) (*stripe.CustomerCashBalanceTransaction, error) {
	params := &stripe.CustomerCashBalanceTransactionParams{
		Customer: stripe.String(cus_id),
	}

	cbt, err := s.client.CustomerCashBalanceTransactions.Get(cbt_id, params)
	if err != nil {
		return nil, err
	}

	return cbt, nil
}

// https://stripe.com/docs/api/cash_balance_transactions/list
// Iterate through cash balance transactions
/*
for i.Next() {
	  cbt := i.CustomerCashBalanceTransaction()
}
*/
func (s *Stripe) ListCustomerCashBalanceTransactions(cus_id string, params *stripe.CustomerCashBalanceTransactionListParams) *customercashbalancetransaction.Iter {
	params.Customer = stripe.String(cus_id)
	i := s.client.CustomerCashBalanceTransactions.List(params)
	return i
}

// https://stripe.com/docs/api/cash_balance_transactions/fund_cash_balance
func (s *Stripe) FundTestModeCashBalance(cus_id, currency_3ISO string, amount int64) (*stripe.CustomerCashBalanceTransaction, error) {
	params := &stripe.
		TestHelpersCustomerFundCashBalanceParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(string(currency_3ISO)),
	}

	cbt, err := s.client.TestHelpersCustomers.FundCashBalance(cus_id, params)
	if err != nil {
		return nil, err
	}

	return cbt, nil
}

//SECTION - CARDS
/*
	You can store multiple cards on a customer in order to charge the customer later. You can also store multiple debit cards on a recipient in order to transfer to those cards later.
*/

// https://stripe.com/docs/api/cards/create
func (s *Stripe) CreatePaymentCard(cus_id string, params *stripe.CardParams) (*stripe.Card, error) {
	params.Customer = stripe.String(cus_id)

	c, err := s.client.Cards.New(params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/cards/retrieve
// card_id is prepended with card_
func (s *Stripe) GetPaymentCard(cus_id, card_id string) (*stripe.Card, error) {
	params := &stripe.CardParams{
		Customer: stripe.String(cus_id),
	}

	c, err := s.client.Cards.Get(card_id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/cards/update
func (s *Stripe) UpdatePaymentCard(cus_id, card_id string, params *stripe.CardParams) (*stripe.Card, error) {
	params.Customer = stripe.String(cus_id)

	c, err := s.client.Cards.Update(card_id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/cards/delete
// card_id is prepended with card_
func (s *Stripe) DeletePaymentCard(cus_id, card_id string) (*stripe.Card, error) {
	params := &stripe.CardParams{
		Customer: stripe.String(cus_id),
	}

	c, err := s.client.Cards.Del(card_id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/cards/list
// Iterate through cards
/*
for i.Next() {
	  c := i.Card()
}
*/
func (s *Stripe) ListPaymentCards(cus_id string, params *stripe.CardListParams) *card.Iter {
	params.Customer = stripe.String(cus_id)

	i := s.client.Cards.List(params)
	return i
}

//SECTION - SOURCES (DEPRECATED)

//SECTION - PRODUCTS
/*
	Products describe the specific goods or services you offer to your customers. For example, you might offer a Standard and Premium version of your goods or service; each version would be a separate Product. They can be used in conjunction with Prices to configure pricing in Payment Links, Checkout, and Subscriptions.
*/

// https://stripe.com/docs/api/products/create
// Required Params - Name
func (s *Stripe) CreateProduct(params *stripe.ProductParams) (*stripe.Product, error) {
	p, err := s.client.Products.New(params)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/products/retrieve
// product_id is prepended with prod_
func (s *Stripe) GetProduct(product_id string) (*stripe.Product, error) {
	p, err := s.client.Products.Get(product_id, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/products/update
func (s *Stripe) UpdateProduct(product_id string, params *stripe.ProductParams) (*stripe.Product, error) {
	p, err := s.client.Products.Update(product_id, params)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/products/list
// Iterate through products
/*
for i.Next() {
	  p := i.Product()
}
*/
func (s *Stripe) ListProducts(params *stripe.ProductListParams) *product.Iter {
	i := s.client.Products.List(params)
	return i
}

// https://stripe.com/docs/api/products/delete
// product_id is prepended with prod_
func (s *Stripe) DeleteProduct(product_id string) (*stripe.Product, error) {
	p, err := s.client.Products.Del(product_id, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/products/search
// Iterate through products
/*
for i.Next() {
	  p := i.Product()
}
*/
func (s *Stripe) SearchProducts(query string) *product.SearchIter {
	params := &stripe.ProductSearchParams{}
	params.Query = *stripe.String(query)
	i := s.client.Products.Search(params)
	return i
}

//SECTION - PRICES
/*
	Prices define the unit cost, currency, and (optional) billing cycle for both recurring and one-time purchases of products. Products help you track inventory or provisioning, and prices help you track payment terms. Different physical goods or levels of service should be represented by products, and pricing options should be represented by prices. This approach lets you change prices without having to change your provisioning scheme.

	For example, you might have a single "gold" product that has prices for $10/month, $100/year, and €9 once.
*/

// https://stripe.com/docs/api/prices/create
// Required Params - UnitAmount, Currency, Recurring.Interval
/*
	- `currency` -
	REQUIRED
	Three-letter ISO currency code, in lowercase. Must be a supported currency.
*/
/*
	- `product` -
	REQUIRED UNLESS PRODUCT_DATA IS PROVIDED
	The ID of the product that this price will belong to.
*/
/*
	- `unit_amount` -
	REQUIRED CONDITIONALLY
	A positive integer in pence (or 0 for a free price) representing how much to charge. One of unit_amount or custom_unit_amount is required, unless billing_scheme=tiered.
*/
func (s *Stripe) CreatePrice(currency string, params *stripe.PriceParams) (*stripe.Price, error) {
	params.Currency = stripe.String(currency)

	p, err := s.client.Prices.New(params)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/prices/retrieve
func (s *Stripe) GetPrice(price_id string) (*stripe.Price, error) {
	p, err := s.client.Prices.Get(price_id, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/prices/update
func (s *Stripe) UpdatePrice(price_id string, params *stripe.PriceParams) (*stripe.Price, error) {
	p, err := s.client.Prices.Update(price_id, params)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/prices/list
// Iterate through prices
/*
for i.Next() {
	  p := i.Price()
}
*/
func (s *Stripe) ListPrices(opt_params *stripe.PriceListParams) *price.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Prices.List(opt_params)
	return i
}

// https://stripe.com/docs/api/prices/search
// Iterate through prices
/*
for i.Next() {
	  p := i.Price()
}
*/
func (s *Stripe) SearchPrices(query string) *price.SearchIter {
	params := &stripe.PriceSearchParams{}
	params.Query = *stripe.String(query)
	i := s.client.Prices.Search(params)
	return i
}

//SECTION - COUPONS
/*
	A coupon contains information about a percent-off or amount-off discount you might want to apply to a customer. Coupons may be applied to subscriptions, invoices, checkout sessions, quotes, and more. Coupons do not work with conventional one-off charges or payment intents.
*/

// https://stripe.com/docs/api/coupons/create
// Required Params - PercentOff or AmountOff
func (s *Stripe) CreateCoupon(params *stripe.CouponParams) (*stripe.Coupon, error) {
	c, err := s.client.Coupons.New(params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/coupons/retrieve
func (s *Stripe) GetCoupon(coupon_id string) (*stripe.Coupon, error) {
	c, err := s.client.Coupons.Get(coupon_id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/coupons/update
func (s *Stripe) UpdateCoupon(coupon_id string, params *stripe.CouponParams) (*stripe.Coupon, error) {
	c, err := s.client.Coupons.Update(coupon_id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/coupons/delete
func (s *Stripe) DeleteCoupon(coupon_id string) (*stripe.Coupon, error) {
	c, err := s.client.Coupons.Del(coupon_id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/coupons/list
// Iterate through coupons
/*
for i.Next() {
	  c := i.Coupon()
}
*/
func (s *Stripe) ListCoupons(opt_params *stripe.CouponListParams) *coupon.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Coupons.List(opt_params)
	return i
}

//SECTION - PROMOTION CODE
/*
	A Promotion Code represents a customer-redeemable code for a coupon. It can be used to create multiple codes for a single coupon.
*/

// https://stripe.com/docs/api/promotion_codes/create
// Required Params - Coupon
func (s *Stripe) CreatePromotionCode(coupon_id string, params *stripe.PromotionCodeParams) (*stripe.PromotionCode, error) {
	params.Coupon = stripe.String(coupon_id)

	pc, err := s.client.PromotionCodes.New(params)
	if err != nil {
		return nil, err
	}

	return pc, nil
}

// https://stripe.com/docs/api/promotion_codes/update
func (s *Stripe) UpdatePromotionCode(promotion_code_id string, params *stripe.PromotionCodeParams) (*stripe.PromotionCode, error) {
	pc, err := s.client.PromotionCodes.Update(promotion_code_id, params)
	if err != nil {
		return nil, err
	}

	return pc, nil
}

// https://stripe.com/docs/api/promotion_codes/retrieve
// promotion_code_id is prepended with promo_
func (s *Stripe) GetPromotionCode(promotion_code_id string) (*stripe.PromotionCode, error) {
	pc, err := s.client.PromotionCodes.Get(promotion_code_id, nil)
	if err != nil {
		return nil, err
	}

	return pc, nil
}

// https://stripe.com/docs/api/promotion_codes/list
// Iterate through promotion codes
/*
for i.Next() {
	  pc := i.PromotionCode()
}
*/
func (s *Stripe) ListPromotionCodes(opt_params *stripe.PromotionCodeListParams) *promotioncode.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.PromotionCodes.List(opt_params)
	return i
}

//SECTION - DISCOUNTS
/*
	A discount represents the actual application of a coupon or promotion code. It contains information about when the discount began, when it will end, and what it is applied to.
*/

// https://stripe.com/docs/api/discounts/delete
// removes a discount from a customer
func (s *Stripe) DeleteDiscount(cus_id string) (*stripe.Customer, error) {
	cus, err := s.client.Customers.DeleteDiscount(cus_id, nil)
	if err != nil {
		return nil, err
	}

	return cus, nil
}

// https://stripe.com/docs/api/discounts/subscription_delete
// removes a discount from a subscription
// subscription_id is prepended with sub_
func (s *Stripe) DeleteSubscriptionDiscount(sub_id string) (*stripe.Subscription, error) {
	d, err := s.client.Subscriptions.DeleteDiscount(sub_id, nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

//SECTION - TAX CODE
/*
	Tax codes classify goods and services for tax purposes.
*/

// https://stripe.com/docs/api/tax_codes/list
// Iterate through tax codes
/*
for i.Next() {
	  tc := i.TaxCode()
}
*/
func (s *Stripe) ListTaxCodes(opt_params *stripe.TaxCodeListParams) *taxcode.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.TaxCodes.List(opt_params)
	return i
}

// https://stripe.com/docs/api/tax_codes/retrieve
// tax_code_id is prepended with txcd_
func (s *Stripe) GetTaxCode(tax_code_id string) (*stripe.TaxCode, error) {
	tc, err := s.client.TaxCodes.Get(tax_code_id, nil)
	if err != nil {
		return nil, err
	}

	return tc, nil
}

//SECTION - TAX RATE
/*
	Tax rates can be applied to invoices, subscriptions and Checkout Sessions to collect tax.
*/

// https://stripe.com/docs/api/tax_rates/create
// Required Params - display_name, inclusive, percentage
func (s *Stripe) CreateTaxRate(display_name string, inclusive bool, percentage float64, params *stripe.TaxRateParams) (*stripe.TaxRate, error) {
	params.DisplayName = stripe.String(display_name)
	params.Inclusive = stripe.Bool(inclusive)
	params.Percentage = stripe.Float64(percentage)

	tr, err := s.client.TaxRates.New(params)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// https://stripe.com/docs/api/tax_rates/retrieve
// tax_rate_id is prepended with txr_
func (s *Stripe) GetTaxRate(tax_rate_id string) (*stripe.TaxRate, error) {
	tr, err := s.client.TaxRates.Get(tax_rate_id, nil)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// https://stripe.com/docs/api/tax_rates/update
func (s *Stripe) UpdateTaxRate(tax_rate_id string, params *stripe.TaxRateParams) (*stripe.TaxRate, error) {
	tr, err := s.client.TaxRates.Update(tax_rate_id, params)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// https://stripe.com/docs/api/tax_rates/list
// Iterate through tax rates
/*
for i.Next() {
	  tr := i.TaxRate()
}
*/
func (s *Stripe) ListTaxRates(opt_params *stripe.TaxRateListParams) *taxrate.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.TaxRates.List(opt_params)
	return i
}

//SECTION - SHIPPING RATES
/*
	Shipping rates describe the price of shipping presented to your customers and applied to a purchase. For more information, see Charge for shipping.
*/

// https://stripe.com/docs/api/shipping_rates/create
// Required Params - display_name
func (s *Stripe) CreateShippingRate(display_name string, params *stripe.ShippingRateParams) (*stripe.ShippingRate, error) {
	params.DisplayName = stripe.String(display_name)
	params.Type = stripe.String("fixed_amount") //This is the only option for now. 10/8/2021

	sr, err := s.client.ShippingRates.New(params)
	if err != nil {
		return nil, err
	}

	return sr, nil
}

// https://stripe.com/docs/api/shipping_rates/retrieve
// shipping_rate_id is prepended with shr_
func (s *Stripe) GetShippingRate(shipping_rate_id string) (*stripe.ShippingRate, error) {
	sr, err := s.client.ShippingRates.Get(shipping_rate_id, nil)
	if err != nil {
		return nil, err
	}

	return sr, nil
}

// https://stripe.com/docs/api/shipping_rates/update
func (s *Stripe) UpdateShippingRate(shipping_rate_id string, params *stripe.ShippingRateParams) (*stripe.ShippingRate, error) {
	sr, err := s.client.ShippingRates.Update(shipping_rate_id, params)
	if err != nil {
		return nil, err
	}

	return sr, nil
}

// https://stripe.com/docs/api/shipping_rates/list
// Iterate through shipping rates
/*
for i.Next() {
	  sr := i.ShippingRate()
}
*/
func (s *Stripe) ListShippingRates(opt_params *stripe.ShippingRateListParams) *shippingrate.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.ShippingRates.List(opt_params)
	return i
}

//SECTION - SESSIONS
/*

	A Checkout Session represents your customer's session as they pay for one-time purchases or subscriptions through Checkout or Payment Links. We recommend creating a new Session each time your customer attempts to pay.

	Once payment is successful, the Checkout Session will contain a reference to the Customer, and either the successful PaymentIntent or an active Subscription.

*/

// https://stripe.com/docs/api/checkout/sessions/create
// Required Params - line_items, mode, success_url
/*
	- `mode` -
	Possible Enum Values: payment, setup, subscription
*/
func (s *Stripe) CreateCheckoutSession(params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	cs, err := s.client.CheckoutSessions.New(params)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// https://stripe.com/docs/api/checkout/sessions/expire
// session_id is prepended with cs_
func (s *Stripe) ExpireCheckoutSession(session_id string) (*stripe.CheckoutSession, error) {
	cs, err := s.client.CheckoutSessions.Expire(session_id, nil)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// https://stripe.com/docs/api/checkout/sessions/retrieve
// session_id is prepended with cs_
func (s *Stripe) GetCheckoutSession(session_id string) (*stripe.CheckoutSession, error) {
	cs, err := s.client.CheckoutSessions.Get(session_id, nil)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// https://stripe.com/docs/api/checkout/sessions/list
// Iterate through checkout sessions
/*
for i.Next() {
	  cs := i.CheckoutSession()
}
*/
func (s *Stripe) ListCheckoutSessions(opt_params *stripe.CheckoutSessionListParams) *session.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.CheckoutSessions.List(opt_params)
	return i
}

// https://stripe.com/docs/api/checkout/sessions/line_items
// Iterate through line items
/*
for i.Next() {
	  li := i.LineItem()
}
*/
func (s *Stripe) ListCheckoutSessionLineItems(session_id string, opt_params *stripe.CheckoutSessionListLineItemsParams) *session.LineItemIter {
	opt_params.Filters.AddFilter("limit", "", "5")
	opt_params.Session = stripe.String(session_id) //TODO: docs conflict TEST THIS
	i := s.client.CheckoutSessions.ListLineItems(opt_params)
	return i
}

//SECTION - PAYMENT LINK
/*
	A payment link is a shareable URL that will take your customers to a hosted payment page. A payment link can be shared and used multiple times.

	When a customer opens a payment link it will open a new checkout session to render the payment page. You can use checkout session events to track payments through payment links.
*/

// TODO - Create another with prefilled options.
// https://stripe.com/docs/api/payment_links/payment_links/create
// Required Params - line_items
func (s *Stripe) CreatePaymentLink(line_items []*stripe.PaymentLinkLineItemParams, params *stripe.PaymentLinkParams) (*stripe.PaymentLink, error) {
	params.LineItems = line_items

	pl, err := s.client.PaymentLinks.New(params)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

// https://stripe.com/docs/api/payment_links/payment_links/retrieve
// payment_link_id is prepended with plink_
func (s *Stripe) GetPaymentLink(payment_link_id string) (*stripe.PaymentLink, error) {
	pl, err := s.client.PaymentLinks.Get(payment_link_id, nil)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

// https://stripe.com/docs/api/payment_links/payment_links/update
func (s *Stripe) UpdatePaymentLink(payment_link_id string, params *stripe.PaymentLinkParams) (*stripe.PaymentLink, error) {
	pl, err := s.client.PaymentLinks.Update(payment_link_id, params)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

// https://stripe.com/docs/api/payment_links/payment_links/list
// Iterate through payment links
/*
for i.Next() {
	  pl := i.PaymentLink()
}
*/
func (s *Stripe) ListPaymentLinks(opt_params *stripe.PaymentLinkListParams) *paymentlink.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.PaymentLinks.List(opt_params)
	return i
}

// https://stripe.com/docs/api/payment_links/line_items
// Iterate through line items
/*
for i.Next() {
	  li := i.LineItem()
}
*/
func (s *Stripe) ListPaymentLinkLineItems(payment_link_id string, opt_params *stripe.PaymentLinkListLineItemsParams) *paymentlink.LineItemIter {
	opt_params.Filters.AddFilter("limit", "", "5")
	opt_params.PaymentLink = stripe.String(payment_link_id) //TODO: docs conflict TEST THIS
	i := s.client.PaymentLinks.ListLineItems(opt_params)
	return i
}

//SECTION - CREDIT NOTE
/*
	Issue a credit note to adjust an invoice's amount after the invoice is finalized.
*/

// https://stripe.com/docs/api/credit_notes/preview
// Required Params - invoice
// invoice_id is prepended with in_
func (s *Stripe) PreviewCreditNote(invoice_id string, params *stripe.CreditNotePreviewParams) (*stripe.CreditNote, error) {
	params.Invoice = stripe.String(invoice_id)

	cn, err := s.client.CreditNotes.Preview(params)
	if err != nil {
		return nil, err
	}

	return cn, nil
}

// https://stripe.com/docs/api/credit_notes/create
// Required Params - invoice
// invoice_id is prepended with in_
func (s *Stripe) CreateCreditNote(invoice_id string, params *stripe.CreditNoteParams) (*stripe.CreditNote, error) {
	params.Invoice = stripe.String(invoice_id)

	cn, err := s.client.CreditNotes.New(params)
	if err != nil {
		return nil, err
	}

	return cn, nil
}

// https://stripe.com/docs/api/credit_notes/retrieve
// credit_note_id is prepended with cn_
func (s *Stripe) GetCreditNote(credit_note_id string) (*stripe.CreditNote, error) {
	cn, err := s.client.CreditNotes.Get(credit_note_id, nil)
	if err != nil {
		return nil, err
	}

	return cn, nil
}

// https://stripe.com/docs/api/credit_notes/update
func (s *Stripe) UpdateCreditNote(credit_note_id string, params *stripe.CreditNoteParams) (*stripe.CreditNote, error) {
	cn, err := s.client.CreditNotes.Update(credit_note_id, params)
	if err != nil {
		return nil, err
	}

	return cn, nil
}

// https://stripe.com/docs/api/credit_notes/lines
// Iterate through credit note line_items
/*
for i.Next() {
	  li := i.LineItem()
}
*/
func (s *Stripe) ListCreditNoteLineItems(credit_note_id string, opt_params *stripe.CreditNoteListLinesParams) *creditnote.LineItemIter {
	opt_params.Filters.AddFilter("limit", "", "5")
	opt_params.CreditNote = stripe.String(credit_note_id)
	i := s.client.CreditNotes.ListLines(opt_params)
	return i
}

// https://stripe.com/docs/api/credit_notes/preview_lines
/*
Required Params - invoice_id -
invoice_id is prepended with in_
*/
// Iterate through credit note line_items
/*
for i.Next() {
	  li := i.LineItem()
}
*/
func (s *Stripe) PreviewCreditNoteLineItems(invoice_id string, opt_params *stripe.CreditNotePreviewLinesParams) *creditnote.LineItemIter {
	opt_params.Invoice = stripe.String(invoice_id)
	i := s.client.CreditNotes.PreviewLines(opt_params)
	return i
}

// https://stripe.com/docs/api/credit_notes/void
// credit_note_id is prepended with cn_
func (s *Stripe) VoidCreditNote(credit_note_id string) (*stripe.CreditNote, error) {
	cn, err := s.client.CreditNotes.VoidCreditNote(credit_note_id, nil)
	if err != nil {
		return nil, err
	}

	return cn, nil
}

// https://stripe.com/docs/api/credit_notes/list
// Iterate through credit notes
/*
for i.Next() {
	  cn := i.CreditNote()
}
*/
func (s *Stripe) ListCreditNotes(opt_params *stripe.CreditNoteListParams) *creditnote.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.CreditNotes.List(opt_params)
	return i
}

//SECTION - CUSTOMER BALANCE TRANSACTION
/*
	Each customer has a Balance value, which denotes a debit or credit that's automatically applied to their next invoice upon finalization. You may modify the value directly by using the update customer API, or by creating a Customer Balance Transaction, which increments or decrements the customer's balance by the specified amount.
*/

// https://stripe.com/docs/api/customer_balance_transactions/create
// Required Params - amount, currency
func (s *Stripe) CreateCustomerBalanceTransaction(cus_id, currency_3ISO string, amount int64, params *stripe.CustomerBalanceTransactionParams) (*stripe.CustomerBalanceTransaction, error) {
	params.Amount = stripe.Int64(amount)
	params.Currency = stripe.String(string(currency_3ISO))
	params.Customer = stripe.String(cus_id)

	cbt, err := s.client.CustomerBalanceTransactions.New(params)
	if err != nil {
		return nil, err
	}

	return cbt, nil
}

// https://stripe.com/docs/api/customer_balance_transactions/retrieve
// cbt_id is prepended with cbtxn_
func (s *Stripe) GetCustomerBalanceTransaction(cbt_id string) (*stripe.CustomerBalanceTransaction, error) {
	cbt, err := s.client.CustomerBalanceTransactions.Get(cbt_id, nil)
	if err != nil {
		return nil, err
	}

	return cbt, nil
}

// https://stripe.com/docs/api/customer_balance_transactions/update
// cbt_id is prepended with cbtxn_
func (s *Stripe) UpdateCustomerBalanceTransaction(cbt_id string, params *stripe.CustomerBalanceTransactionParams) (*stripe.CustomerBalanceTransaction, error) {
	cbt, err := s.client.CustomerBalanceTransactions.Update(cbt_id, params)
	if err != nil {
		return nil, err
	}

	return cbt, nil
}

// https://stripe.com/docs/api/customer_balance_transactions/list
// Iterate through customer balance transactions
/*
for i.Next() {
	  cbt := i.CustomerBalanceTransaction()
}
*/
func (s *Stripe) ListCustomerBalanceTransactions(opt_params *stripe.CustomerBalanceTransactionListParams) *customerbalancetransaction.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.CustomerBalanceTransactions.List(opt_params)
	return i
}

//SECTION - CUSTOMER PORTAL
/*
	The Billing customer portal is a Stripe-hosted UI for subscription and billing management.

	A portal configuration describes the functionality and features that you want to provide to your customers through the portal.

	A portal session describes the instantiation of the customer portal for a particular customer. By visiting the session's URL, the customer can manage their subscriptions and billing details. For security reasons, sessions are short-lived and will expire if the customer does not visit the URL. Create sessions on-demand when customers intend to manage their subscriptions and billing details.
*/

// https://stripe.com/docs/api/customer_portal/sessions/create
// Required Params - customer
// customer_id is prepended with cus_
func (s *Stripe) CreateCustomerPortalSession(customer_id string, params *stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
	params.Customer = stripe.String(customer_id)

	bps, err := s.client.BillingPortalSessions.New(params)
	if err != nil {
		return nil, err
	}

	return bps, nil
}

// https://stripe.com/docs/api/customer_portal/configurations/create
// Required Params - business_profile, features
func (s *Stripe) CreateCustomerPortalConfiguration(features *stripe.BillingPortalConfigurationFeaturesParams, business_profile *stripe.BillingPortalConfigurationBusinessProfileParams, params *stripe.BillingPortalConfigurationParams) (*stripe.BillingPortalConfiguration, error) {
	params.Features = features
	params.BusinessProfile = business_profile

	bpc, err := s.client.BillingPortalConfigurations.New(params)
	if err != nil {
		return nil, err
	}

	return bpc, nil
}

// https://stripe.com/docs/api/customer_portal/configurations/update
// bpc_id is prepended with bpc_
func (s *Stripe) UpdateCustomerPortalConfiguration(bpc_id string, params *stripe.BillingPortalConfigurationParams) (*stripe.BillingPortalConfiguration, error) {
	bpc, err := s.client.BillingPortalConfigurations.Update(bpc_id, params)
	if err != nil {
		return nil, err
	}

	return bpc, nil
}

// https://stripe.com/docs/api/customer_portal/configurations/retrieve
// bpc_id is prepended with bpc_
func (s *Stripe) GetCustomerPortalConfiguration(bpc_id string) (*stripe.BillingPortalConfiguration, error) {
	bpc, err := s.client.BillingPortalConfigurations.Get(bpc_id, nil)
	if err != nil {
		return nil, err
	}

	return bpc, nil
}

// https://stripe.com/docs/api/customer_portal/configurations/list
// Iterate through customer portal configurations
/*
for i.Next() {
	  bpc := i.BillingPortalConfiguration()
}
*/
func (s *Stripe) ListCustomerPortalConfigurations(opt_params *stripe.BillingPortalConfigurationListParams) *bpc.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.BillingPortalConfigurations.List(opt_params)
	return i
}

//SECTION - CUSTOMER TAX IDS
/*
	You can add one or multiple tax IDs to a customer. A customer's tax IDs are displayed on invoices and credit notes issued for the customer.
*/

// https://stripe.com/docs/api/customer_tax_ids/create
// Required Params - type, value
/* - `tax_type` -
- ad_nrt, ae_trn, ar_cuit, au_abn, au_arn, bg_uic, bo_tin, br_cnpj, br_cpf, ca_bn, ca_gst_hst, ca_pst_bc, ca_pst_mb, ca_pst_sk, ca_qst, ch_vat, cl_tin, cn_tin, co_nit, cr_tin, do_rcn, ec_ruc, eg_tin, es_cif, eu_oss_vat, eu_vat, gb_vat, ge_vat, hk_br, hu_tin, id_npwp, il_vat, in_gst, is_vat, jp_cn, jp_rn, jp_trn, ke_pin, kr_brn, li_uid, mx_rfc, my_frp, my_itn, my_sst, no_vat, nz_gst, pe_ruc, ph_tin, ro_tin, rs_pib, ru_inn, ru_kpp, sa_vat, sg_gst, sg_uen, si_tin, sv_nit, th_vat, tr_tin, tw_vat, ua_vat, us_ein, uy_ruc, ve_rif, vn_tin, or za_vat
*/
// customer_id is prepended with cus_
func (s *Stripe) CreateTaxIDForCustomer(customer_id, tax_type, tax_id string) (*stripe.TaxID, error) {
	params := &stripe.TaxIDParams{
		Type:     stripe.String(tax_type),
		Value:    stripe.String(tax_id),
		Customer: stripe.String(customer_id),
	}

	cti, err := s.client.TaxIDs.New(params)
	if err != nil {
		return nil, err
	}

	return cti, nil
}

// https://stripe.com/docs/api/customer_tax_ids/create
// Required Params - type, value
/* - `tax_type` -
- ad_nrt, ae_trn, ar_cuit, au_abn, au_arn, bg_uic, bo_tin, br_cnpj, br_cpf, ca_bn, ca_gst_hst, ca_pst_bc, ca_pst_mb, ca_pst_sk, ca_qst, ch_vat, cl_tin, cn_tin, co_nit, cr_tin, do_rcn, ec_ruc, eg_tin, es_cif, eu_oss_vat, eu_vat, gb_vat, ge_vat, hk_br, hu_tin, id_npwp, il_vat, in_gst, is_vat, jp_cn, jp_rn, jp_trn, ke_pin, kr_brn, li_uid, mx_rfc, my_frp, my_itn, my_sst, no_vat, nz_gst, pe_ruc, ph_tin, ro_tin, rs_pib, ru_inn, ru_kpp, sa_vat, sg_gst, sg_uen, si_tin, sv_nit, th_vat, tr_tin, tw_vat, ua_vat, us_ein, uy_ruc, ve_rif, vn_tin, or za_vat
*/
func (s *Stripe) CreateTaxID(tax_type, tax_id string) (*stripe.TaxID, error) {
	params := &stripe.TaxIDParams{
		Type:  stripe.String(tax_type),
		Value: stripe.String(tax_id),
	}

	cti, err := s.client.TaxIDs.New(params)
	if err != nil {
		return nil, err
	}

	return cti, nil
}

// https://stripe.com/docs/api/customer_tax_ids/retrieve
// tax_id is prepended with txi_
func (s *Stripe) GetTaxID(tax_id string) (*stripe.TaxID, error) {
	cti, err := s.client.TaxIDs.Get(tax_id, nil)
	if err != nil {
		return nil, err
	}

	return cti, nil
}

// https://stripe.com/docs/api/customer_tax_ids/delete
// tax_id is prepended with txi_
func (s *Stripe) DeleteTaxID(tax_id string) (*stripe.TaxID, error) {
	cti, err := s.client.TaxIDs.Del(tax_id, nil)
	if err != nil {
		return nil, err
	}

	return cti, nil
}

// https://stripe.com/docs/api/customer_tax_ids/list
// Iterate through tax ids
/*
for i.Next() {
	  cti := i.TaxID()
}
*/
func (s *Stripe) ListTaxIDs(opt_params *stripe.TaxIDListParams) *taxid.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.TaxIDs.List(opt_params)
	return i
}

//SECTION - INVOICES
/*
	Invoices are statements of amounts owed by a customer, and are either generated one-off, or generated periodically from a subscription.

	They contain invoice items, and proration adjustments that may be caused by subscription upgrades/downgrades (if necessary).

	If your invoice is configured to be billed through automatic charges, Stripe automatically finalizes your invoice and attempts payment. Note that finalizing the invoice, when automatic, does not happen immediately as the invoice is created. Stripe waits until one hour after the last webhook was successfully sent (or the last webhook timed out after failing). If you (and the platforms you may have connected to) have no webhooks configured, Stripe waits one hour after creation to finalize the invoice.

	If your invoice is configured to be billed by sending an email, then based on your email settings, Stripe will email the invoice to your customer and await payment. These emails can contain a link to a hosted page to pay the invoice.

	Stripe applies any customer credit on the account before determining the amount due for the invoice (i.e., the amount that will be actually charged). If the amount due for the invoice is less than Stripe's minimum allowed charge per currency, the invoice is automatically marked paid, and we add the amount due to the customer's credit balance which is applied to the next invoice.
*/

// https://stripe.com/docs/api/invoices/create
func (s *Stripe) CreateInvoice(params *stripe.InvoiceParams) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.New(params)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/retrieve
// invoice_id is prepended with in_
func (s *Stripe) GetInvoice(invoice_id string) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.Get(invoice_id, nil)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/update
// invoice_id is prepended with in_
func (s *Stripe) UpdateInvoice(invoice_id string, params *stripe.InvoiceParams) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.Update(invoice_id, params)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/delete
// invoice_id is prepended with in_
func (s *Stripe) DeleteDraftInvoice(invoice_id string) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.Del(invoice_id, nil)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/finalize
// invoice_id is prepended with in_
func (s *Stripe) FinalizeInvoice(invoice_id string) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.FinalizeInvoice(invoice_id, nil)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/pay
// invoice_id is prepended with in_
func (s *Stripe) PayInvoice(invoice_id string, params *stripe.InvoicePayParams) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.Pay(invoice_id, params)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/send
// invoice_id is prepended with in_
func (s *Stripe) SendInvoiceForManualPayment(invoice_id string, params *stripe.InvoiceSendInvoiceParams) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.SendInvoice(invoice_id, params)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/void
// invoice_id is prepended with in_
func (s *Stripe) VoidInvoice(invoice_id string) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.VoidInvoice(invoice_id, nil)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/mark_uncollectible
// invoice_id is prepended with in_
func (s *Stripe) MarkInvoiceUncollectible(invoice_id string) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.MarkUncollectible(invoice_id, nil)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/invoice_lines
// Iterate through invoice lines
/*
for i.Next() {
	  il := i.InvoiceLine()
}
*/
func (s *Stripe) ListInvoiceLines(invoice_id string, opt_params *stripe.InvoiceListLinesParams) *invoice.LineItemIter {
	opt_params.Filters.AddFilter("limit", "", "3")
	opt_params.Invoice = stripe.String(invoice_id)
	i := s.client.Invoices.ListLines(opt_params)
	return i
}

// https://stripe.com/docs/api/invoices/upcoming
func (s *Stripe) GetUpcomingInvoice(params *stripe.InvoiceUpcomingParams) (*stripe.Invoice, error) {
	i, err := s.client.Invoices.Upcoming(params)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// https://stripe.com/docs/api/invoices/upcoming_invoice_lines
// Iterate through invoice lines
/*
for i.Next() {
	  il := i.InvoiceLine()
}
*/
func (s *Stripe) ListUpcomingInvoiceLines(opt_params *stripe.InvoiceUpcomingLinesParams) *invoice.LineItemIter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Invoices.UpcomingLines(opt_params)
	return i
}

// https://stripe.com/docs/api/invoices/list
// Iterate through invoices
/*
for i.Next() {
	  i := i.Invoice()
}
*/
func (s *Stripe) ListInvoices(opt_params *stripe.InvoiceListParams) *invoice.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Invoices.List(opt_params)
	return i
}

// https://stripe.com/docs/api/invoices/search
// Iterate through invoices
/*
for i.Next() {
	  i := i.Invoice()
}
*/
func (s *Stripe) SearchInvoices(query string) *invoice.SearchIter {
	params := &stripe.InvoiceSearchParams{}
	params.Query = *stripe.String(query)

	i := s.client.Invoices.Search(params)
	return i
}

//SECTION - INVOICE ITEMS
/*
	Invoice Items represent the component lines of an invoice. An invoice item is added to an invoice by creating or updating it with an invoice field, at which point it will be included as an invoice line item within invoice.lines.

	Invoice Items can be created before you are ready to actually send the invoice. This can be particularly useful when combined with a subscription. Sometimes you want to add a charge or credit to a customer, but actually charge or credit the customer’s card only at the end of a regular billing cycle. This is useful for combining several charges (to minimize per-transaction fees), or for having Stripe tabulate your usage-based billing totals.
*/

// https://stripe.com/docs/api/invoiceitems/create
// Required Params - customer
// customer_id is prepended with cus_
func (s *Stripe) CreateInvoiceItem(customer_id string, params *stripe.InvoiceItemParams) (*stripe.InvoiceItem, error) {
	params.Customer = stripe.String(customer_id)

	ii, err := s.client.InvoiceItems.New(params)
	if err != nil {
		return nil, err
	}

	return ii, nil
}

// https://stripe.com/docs/api/invoiceitems/retrieve
// invoice_item_id is prepended with ii_
func (s *Stripe) GetInvoiceItem(invoice_item_id string) (*stripe.InvoiceItem, error) {
	ii, err := s.client.InvoiceItems.Get(invoice_item_id, nil)
	if err != nil {
		return nil, err
	}

	return ii, nil
}

// https://stripe.com/docs/api/invoiceitems/update
// invoice_item_id is prepended with ii_
func (s *Stripe) UpdateInvoiceItem(invoice_item_id string, params *stripe.InvoiceItemParams) (*stripe.InvoiceItem, error) {
	ii, err := s.client.InvoiceItems.Update(invoice_item_id, params)
	if err != nil {
		return nil, err
	}

	return ii, nil
}

// https://stripe.com/docs/api/invoiceitems/delete
// invoice_item_id is prepended with ii_
func (s *Stripe) DeleteInvoiceItem(invoice_item_id string) (*stripe.InvoiceItem, error) {
	ii, err := s.client.InvoiceItems.Del(invoice_item_id, nil)
	if err != nil {
		return nil, err
	}

	return ii, nil
}

// https://stripe.com/docs/api/invoiceitems/list
// Iterate through invoice items
/*
for i.Next() {
	  ii := i.InvoiceItem()
}
*/
func (s *Stripe) ListInvoiceItems(opt_params *stripe.InvoiceItemListParams) *invoiceitem.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.InvoiceItems.List(opt_params)
	return i
}

//SECTION - PLANS
// subscription plans, not implemented as can be done with the Prices API

//SECTION - QUOTE
/*
	A Quote is a way to model prices that you'd like to provide to a customer. Once accepted, it will automatically create an invoice, subscription or subscription schedule.
*/

// https://stripe.com/docs/api/quotes/object
func (s *Stripe) CreateQuote(params *stripe.QuoteParams) (*stripe.Quote, error) {
	q, err := s.client.Quotes.New(params)
	if err != nil {
		return nil, err
	}

	return q, nil
}

// https://stripe.com/docs/api/quotes/retrieve
// quote_id is prepended with qt_
func (s *Stripe) GetQuote(quote_id string) (*stripe.Quote, error) {
	q, err := s.client.Quotes.Get(quote_id, nil)
	if err != nil {
		return nil, err
	}

	return q, nil
}

// https://stripe.com/docs/api/quotes/update
// quote_id is prepended with qt_
func (s *Stripe) UpdateQuote(quote_id string, params *stripe.QuoteParams) (*stripe.Quote, error) {
	q, err := s.client.Quotes.Update(quote_id, params)
	if err != nil {
		return nil, err
	}

	return q, nil
}

// https://stripe.com/docs/api/quotes/finalize
// quote_id is prepended with qt_
func (s *Stripe) FinalizeQuote(quote_id string, opt_params *stripe.QuoteFinalizeQuoteParams) (*stripe.Quote, error) {
	q, err := s.client.Quotes.FinalizeQuote(quote_id, opt_params)
	if err != nil {
		return nil, err
	}

	return q, nil
}

// https://stripe.com/docs/api/quotes/accept
// quote_id is prepended with qt_
func (s *Stripe) AcceptQuote(quote_id string) (*stripe.Quote, error) {
	q, err := s.client.Quotes.Accept(quote_id, nil)
	if err != nil {
		return nil, err
	}

	return q, nil
}

// https://stripe.com/docs/api/quotes/cancel
// quote_id is prepended with qt_
func (s *Stripe) CancelQuote(quote_id string) (*stripe.Quote, error) {
	q, err := s.client.Quotes.Cancel(quote_id, nil)
	if err != nil {
		return nil, err
	}

	return q, nil
}

// https://stripe.com/docs/api/quotes/pdf
// quote_id is prepended with qt_
// TODO: TEST THIS
func (s *Stripe) DownloadQuotePDF(quote_id string, dest_file string) (bool, error) {
	success := false
	q, err := s.client.Quotes.PDF(quote_id, nil)
	if err != nil {
		return success, err
	}

	_ = bufio.NewReader(q.LastResponse.Body)
	defer q.LastResponse.Body.Close()

	f, err := os.Create(dest_file)
	if err != nil {
		return success, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.Flush()

	success = true

	return success, nil
}

// https://stripe.com/docs/api/quotes/line_items/list
// Iterate through quote line items
/*
for i.Next() {
	  li := i.LineItem()
}
*/
func (s *Stripe) ListQuoteLineItems(quote_id string, opt_params *stripe.QuoteListLineItemsParams) *quote.LineItemIter {
	opt_params.Filters.AddFilter("limit", "", "3")
	opt_params.Quote = stripe.String(quote_id)
	i := s.client.Quotes.ListLineItems(opt_params)
	return i
}

// https://stripe.com/docs/api/quotes/line_items/upfront/list
// Iterate through quote line items
/*
for i.Next() {
	  li := i.LineItem()
}
*/
func (s *Stripe) ListUpfrontQuoteLineItems(quote_id string, opt_params *stripe.QuoteListComputedUpfrontLineItemsParams) *quote.LineItemIter {
	opt_params.Filters.AddFilter("limit", "", "3")
	opt_params.Quote = stripe.String(quote_id)
	i := s.client.Quotes.ListComputedUpfrontLineItems(opt_params)
	return i
}

// https://stripe.com/docs/api/quotes/list
// Iterate through quotes
/*
for i.Next() {
	  q := i.Quote()
}
*/
func (s *Stripe) ListQuotes(opt_params *stripe.QuoteListParams) *quote.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Quotes.List(opt_params)
	return i
}

//SECTION - SUBSCRIPTIONS
/*
	Subscriptions allow you to charge a customer on a recurring basis.
*/

// https://stripe.com/docs/api/subscriptions/create
// Required Params - customer, items
// customer_id is prepended with cus_
func (s *Stripe) CreateSubscription(customer_id string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	params.Customer = stripe.String(customer_id)

	if params.Items != nil {
		return nil, errors.New("items are required")
	}

	sub, err := s.client.Subscriptions.New(params)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// https://stripe.com/docs/api/subscriptions/retrieve
// subscription_id is prepended with sub_
func (s *Stripe) GetSubscription(subscription_id string) (*stripe.Subscription, error) {
	sub, err := s.client.Subscriptions.Get(subscription_id, nil)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// https://stripe.com/docs/api/subscriptions/update
// subscription_id is prepended with sub_
func (s *Stripe) UpdateSubscription(subscription_id string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	sub, err := s.client.Subscriptions.Update(subscription_id, params)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// https://stripe.com/docs/api/subscriptions/update
// subscription_id is prepended with sub_
func (s *Stripe) ResumeSubscription(subscription_id string, params *stripe.SubscriptionResumeParams) (*stripe.Subscription, error) {
	sub, err := s.client.Subscriptions.Resume(subscription_id, params)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// https://stripe.com/docs/api/subscriptions/cancel
// subscription_id is prepended with sub_
func (s *Stripe) CancelSubscription(subscription_id string, params *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
	sub, err := s.client.Subscriptions.Cancel(subscription_id, params)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// https://stripe.com/docs/api/subscriptions/list
// Iterate through subscriptions
/*
for i.Next() {
	  sub := i.Subscription()
}
*/
func (s *Stripe) ListSubscriptions(opt_params *stripe.SubscriptionListParams) *subscription.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Subscriptions.List(opt_params)
	return i
}

// https://stripe.com/docs/api/subscriptions/search
// Iterate through subscriptions
/*
for i.Next() {
	  sub := i.Subscription()
}
*/
func (s *Stripe) SearchSubscriptions(query string) *subscription.SearchIter {
	params := &stripe.SubscriptionSearchParams{}
	params.Query = *stripe.String(query)

	i := s.client.Subscriptions.Search(params)
	return i
}

//SECTION - SUBSCRIPTION ITEMS
/*
	Subscription items allow you to create customer subscriptions with more than one plan, making it easy to represent complex billing relationships.
*/

// https://stripe.com/docs/api/subscription_items/create
// Required Params - subscription
// subscription_id is prepended with sub_
func (s *Stripe) CreateSubscriptionItem(subscription_id string, params *stripe.SubscriptionItemParams) (*stripe.SubscriptionItem, error) {
	params.Subscription = stripe.String(subscription_id)

	si, err := s.client.SubscriptionItems.New(params)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/subscription_items/retrieve
// subscription_item_id is prepended with si_
func (s *Stripe) GetSubscriptionItem(subscription_item_id string) (*stripe.SubscriptionItem, error) {
	si, err := s.client.SubscriptionItems.Get(subscription_item_id, nil)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/subscription_items/update
// subscription_item_id is prepended with si_
func (s *Stripe) UpdateSubscriptionItem(subscription_item_id string, params *stripe.SubscriptionItemParams) (*stripe.SubscriptionItem, error) {
	si, err := s.client.SubscriptionItems.Update(subscription_item_id, params)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/subscription_items/delete
// subscription_item_id is prepended with si_
func (s *Stripe) DeleteSubscriptionItem(subscription_item_id string) (*stripe.SubscriptionItem, error) {
	si, err := s.client.SubscriptionItems.Del(subscription_item_id, nil)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// https://stripe.com/docs/api/subscription_items/list
// Iterate through subscription items
/*
for i.Next() {
	  si := i.SubscriptionItem()
}
*/
func (s *Stripe) ListSubscriptionItems(opt_params *stripe.SubscriptionItemListParams) *subscriptionitem.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.SubscriptionItems.List(opt_params)
	return i
}

//SECTION - SUBSCRIPTION SCHEDULE
/*
	A subscription schedule allows you to create and manage the lifecycle of a subscription by predefining expected changes.
*/

// https://stripe.com/docs/api/subscription_schedules/create
func (s *Stripe) CreateSubscriptionSchedule(params *stripe.SubscriptionScheduleParams) (*stripe.SubscriptionSchedule, error) {
	ss, err := s.client.SubscriptionSchedules.New(params)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// https://stripe.com/docs/api/subscription_schedules/retrieve
// subscription_schedule_id is prepended with sub_sched_
func (s *Stripe) GetSubscriptionSchedule(subscription_schedule_id string) (*stripe.SubscriptionSchedule, error) {
	ss, err := s.client.SubscriptionSchedules.Get(subscription_schedule_id, nil)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// https://stripe.com/docs/api/subscription_schedules/update
// subscription_schedule_id is prepended with sub_sched_
func (s *Stripe) UpdateSubscriptionSchedule(subscription_schedule_id string, params *stripe.SubscriptionScheduleParams) (*stripe.SubscriptionSchedule, error) {
	ss, err := s.client.SubscriptionSchedules.Update(subscription_schedule_id, params)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// https://stripe.com/docs/api/subscription_schedules/cancel
// subscription_schedule_id is prepended with sub_sched_
func (s *Stripe) CancelSubscriptionSchedule(subscription_schedule_id string, params *stripe.SubscriptionScheduleCancelParams) (*stripe.SubscriptionSchedule, error) {
	ss, err := s.client.SubscriptionSchedules.Cancel(subscription_schedule_id, params)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// https://stripe.com/docs/api/subscription_schedules/release
// subscription_schedule_id is prepended with sub_sched_
func (s *Stripe) ReleaseSubscriptionSchedule(subscription_schedule_id string, params *stripe.SubscriptionScheduleReleaseParams) (*stripe.SubscriptionSchedule, error) {
	ss, err := s.client.SubscriptionSchedules.Release(subscription_schedule_id, params)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// https://stripe.com/docs/api/subscription_schedules/list
// Iterate through subscription schedules
/*
for i.Next() {
	  ss := i.SubscriptionSchedule()
}
*/
func (s *Stripe) ListSubscriptionSchedules(opt_params *stripe.SubscriptionScheduleListParams) *subscriptionschedule.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.SubscriptionSchedules.List(opt_params)
	return i
}

//SECTION - TEST CLOCKS
/*
	A test clock enables deterministic control over objects in testmode. With a test clock, you can create objects at a frozen time in the past or future, and advance to a specific future time to observe webhooks and state changes. After the clock advances, you can either validate the current state of your scenario (and test your assumptions), change the current state of your scenario (and test more complex scenarios), or keep advancing forward in time.
*/

// https://stripe.com/docs/api/test_clocks/create
// Required Params - frozen_time
func (s *Stripe) CreateTestClock(frozen_time int64, params *stripe.TestHelpersTestClockParams) (*stripe.TestHelpersTestClock, error) {
	params.FrozenTime = stripe.Int64(frozen_time)

	tc, err := s.client.TestHelpersTestClocks.New(params)
	if err != nil {
		return nil, err
	}

	return tc, nil
}

// https://stripe.com/docs/api/test_clocks/retrieve
// test_clock_id is prepended with clock_
func (s *Stripe) GetTestClock(test_clock_id string) (*stripe.TestHelpersTestClock, error) {
	tc, err := s.client.TestHelpersTestClocks.Get(test_clock_id, nil)
	if err != nil {
		return nil, err
	}

	return tc, nil
}

// https://stripe.com/docs/api/test_clocks/delete
// test_clock_id is prepended with clock_
func (s *Stripe) DeleteTestClock(test_clock_id string) (*stripe.TestHelpersTestClock, error) {
	tc, err := s.client.TestHelpersTestClocks.Del(test_clock_id, nil)
	if err != nil {
		return nil, err
	}

	return tc, nil
}

// https://stripe.com/docs/api/test_clocks/advance
// test_clock_id is prepended with clock_
// Required Params - frozen_time
func (s *Stripe) AdvanceTestClock(test_clock_id string, frozen_time int64, params *stripe.TestHelpersTestClockAdvanceParams) (*stripe.TestHelpersTestClock, error) {
	params.FrozenTime = stripe.Int64(frozen_time)
	tc, err := s.client.TestHelpersTestClocks.Advance(test_clock_id, params)
	if err != nil {
		return nil, err
	}

	return tc, nil
}

// https://stripe.com/docs/api/test_clocks/list
// Iterate through test clocks
/*
for i.Next() {
	  tc := i.TestHelpersTestClock()
}
*/
func (s *Stripe) ListTestClocks(opt_params *stripe.TestHelpersTestClockListParams) *testclock.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.TestHelpersTestClocks.List(opt_params)
	return i
}

//SECTION - USAGE RECORDS
/*
	Usage records allow you to report customer usage and metrics to Stripe for metered billing of subscription prices.
*/

// https://stripe.com/docs/api/usage_records/create
// Required Params - quantity, subscription_item
// timestamp is now, if not specified
func (s *Stripe) CreateUsageRecord(subscription_item_id string, quantity int64, params *stripe.UsageRecordParams) (*stripe.UsageRecord, error) {
	params.Quantity = stripe.Int64(quantity)
	params.SubscriptionItem = stripe.String(subscription_item_id)

	ur, err := s.client.UsageRecords.New(params)
	if err != nil {
		return nil, err
	}

	return ur, nil
}

// https://stripe.com/docs/api/usage_records/subscription_item_summary_list
// Iterate through usage records
/*
for i.Next() {
	  ur := i.UsageRecordSummary()
}
*/
func (s *Stripe) ListUsageRecordSummaries(subscription_item_id string, opt_params *stripe.UsageRecordSummaryListParams) *usagerecordsummary.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	opt_params.SubscriptionItem = stripe.String(subscription_item_id)
	i := s.client.UsageRecordSummaries.List(opt_params)
	return i
}

// --- CONNECT ---
//SECTION - ACCOUNTS
/*
	This is an object representing a Stripe account. You can retrieve it to see properties on the account like its current requirements or if the account is enabled to make live charges or receive payouts.

	For Custom accounts, the properties below are always returned. For other accounts, some properties are returned until that account has started to go through Connect Onboarding. Once you create an Account Link for a Standard or Express account, some parameters are no longer returned. These are marked as Custom Only or Custom and Express below. Learn about the differences between accounts.
*/

// https://stripe.com/docs/api/accounts/create
// Required Params - `account_type`, (& `capabilities` for custom accounts)
/*
	- account_type - custom, express, standard
*/
func (s *Stripe) CreateAccount(account_type string, params *stripe.AccountParams) (*stripe.Account, error) {
	params.Type = stripe.String(account_type)
	a, err := s.client.Accounts.New(params)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/accounts/retrieve
// account_id is prepended with acct_
func (s *Stripe) GetAccount(account_id string) (*stripe.Account, error) {
	a, err := s.client.Accounts.GetByID(account_id, nil)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/accounts/update
// account_id is prepended with acct_
func (s *Stripe) UpdateAccount(account_id string, params *stripe.AccountParams) (*stripe.Account, error) {
	a, err := s.client.Accounts.Update(account_id, params)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/accounts/delete
// account_id is prepended with acct_
func (s *Stripe) DeleteAccount(account_id string) (*stripe.Account, error) {
	a, err := s.client.Accounts.Del(account_id, nil)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/account/reject
// account_id is prepended with acct_
// Required - reason
func (s *Stripe) RejectAccount(account_id string, params *stripe.AccountRejectParams) (*stripe.Account, error) {
	a, err := s.client.Accounts.Reject(account_id, params)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/accounts/list
// Iterate through accounts
/*
for i.Next() {
	  a := i.Account()
}
*/
func (s *Stripe) ListAccounts(opt_params *stripe.AccountListParams) *account.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Accounts.List(opt_params)
	return i
}

// https://stripe.com/docs/api/account/create_login_link
// account_id is prepended with acct_
func (s *Stripe) CreateAccountLoginLink(account_id string) (*stripe.LoginLink, error) {
	params := &stripe.LoginLinkParams{
		Account: stripe.String(account_id),
	}

	ll, err := s.client.LoginLinks.New(params)
	if err != nil {
		return nil, err
	}

	return ll, nil
}

//SECTION - ACCOUNT LINKS
/*
	Account Links are the means by which a Connect platform grants a connected account permission to access Stripe-hosted applications, such as Connect Onboarding.
*/

// https://stripe.com/docs/api/account_links/create
// Required Params - account, refresh_url, return_url, type
/*
 - `type` - account_onboarding, account_update
*/
func (s *Stripe) CreateAccountLink(account_id, refresh_url, return_url string, params *stripe.AccountLinkParams) (*stripe.AccountLink, error) {
	params.Account = stripe.String(account_id)
	params.RefreshURL = stripe.String(refresh_url)
	params.ReturnURL = stripe.String(return_url)

	al, err := s.client.AccountLinks.New(params)
	if err != nil {
		return nil, err
	}

	return al, nil
}

//SECTION - APPLICATION FEES
/*
	When you collect a transaction fee on top of a charge made for your user (using Connect), an Application Fee object is created in your account. You can list, retrieve, and refund application fees.
*/

// https://stripe.com/docs/api/application_fees/retrieve
// application_fee_id is prepended with fee_
func (s *Stripe) GetApplicationFee(application_fee_id string) (*stripe.ApplicationFee, error) {
	f, err := s.client.ApplicationFees.Get(application_fee_id, nil)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// https://stripe.com/docs/api/application_fees/list
// Iterate through application fees
/*
for i.Next() {
	  f := i.ApplicationFee()
}
*/
func (s *Stripe) ListApplicationFees(opt_params *stripe.ApplicationFeeListParams) *applicationfee.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.ApplicationFees.List(opt_params)
	return i
}

//SECTION - APPLICATION FEE REFUNDS
/*
	Application Fee Refund objects allow you to refund an application fee that has previously been created but not yet refunded. Funds will be refunded to the Stripe account from which the fee was originally collected.
*/

// https://stripe.com/docs/api/fee_refunds/create
func (s *Stripe) CreateApplicationFeeRefund(application_fee_id string, params *stripe.FeeRefundParams) (*stripe.FeeRefund, error) {
	params.Fee = stripe.String(application_fee_id)

	fr, err := s.client.FeeRefunds.New(params)
	if err != nil {
		return nil, err
	}

	return fr, nil
}

// https://stripe.com/docs/api/fee_refunds/retrieve
// fee_refund_id is prepended with fr_
func (s *Stripe) GetApplicationFeeRefund(fee_refund_id string) (*stripe.FeeRefund, error) {
	fr, err := s.client.FeeRefunds.Get(fee_refund_id, nil)
	if err != nil {
		return nil, err
	}

	return fr, nil
}

// https://stripe.com/docs/api/fee_refunds/update
// fee_refund_id is prepended with fr_
func (s *Stripe) UpdateApplicationFeeRefund(fee_refund_id string, params *stripe.FeeRefundParams) (*stripe.FeeRefund, error) {
	fr, err := s.client.FeeRefunds.Update(fee_refund_id, params)
	if err != nil {
		return nil, err
	}

	return fr, nil
}

// https://stripe.com/docs/api/fee_refunds/list
// Iterate through application fee refunds
/*
for i.Next() {
	  fr := i.FeeRefund()
}
*/
func (s *Stripe) ListApplicationFeeRefunds(opt_params *stripe.FeeRefundListParams) *feerefund.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.FeeRefunds.List(opt_params)
	return i
}

//SECTION - CAPABILITIES
/*
	This is an object representing a capability for a Stripe account.
*/

// https://stripe.com/docs/api/capabilities/retrieve
// account_id is prepended with acct_
func (s *Stripe) GetAccountCapability(capability_id string) (*stripe.Capability, error) {
	c, err := s.client.Capabilities.Get(capability_id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/capabilities/update
// account_id is prepended with acct_
func (s *Stripe) UpdateAccountCapability(capability_id string, params *stripe.CapabilityParams) (*stripe.Capability, error) {
	params.Account = stripe.String(capability_id)

	c, err := s.client.Capabilities.Update(capability_id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/capabilities/list
// Iterate through capabilities
/*
for i.Next() {
	  c := i.Capability()
}
*/
func (s *Stripe) ListAccountCapabilities(opt_params *stripe.CapabilityListParams) *capability.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Capabilities.List(opt_params)
	return i
}

//SECTION - COUNTRY SPECS
/*
	Stripe needs to collect certain pieces of information about each account created. These requirements can differ depending on the account's country. The Country Specs API makes these rules available to your integration.
*/

// https://stripe.com/docs/api/country_specs/list
// Iterate through country specs
/*
for i.Next() {
	  cs := i.CountrySpec()
}
*/
func (s *Stripe) ListCountrySpecs(opt_params *stripe.CountrySpecListParams) *countryspec.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.CountrySpecs.List(opt_params)
	return i
}

// https://stripe.com/docs/api/country_specs/retrieve
func (s *Stripe) GetCountrySpec(country string) (*stripe.CountrySpec, error) {
	cs, err := s.client.CountrySpecs.Get(country, nil)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

//SECTION - EXTERNAL ACCOUNTS
/*
	External Accounts are transfer destinations on Account objects for connected accounts. They can be bank accounts or debit cards.

	Bank accounts and debit cards can also be used as payment sources on regular charges, and are documented in the links above.
*/

//IDENTICAL TO CREATECUSTOMERBANKACCOUNT
// https://stripe.com/docs/api/external_account_bank_accounts/create
// Required Params - external_account
/*
 external_account - bank_account stripe token (acct_...) or a dictionary containing a user's bank account details
*/
// func (s *Stripe) CreateExternalAccountBankAccount(params *stripe.BankAccountParams) (*stripe.BankAccount, error) {
// 	ba, err := s.client.BankAccounts.New(params)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return ba, nil
// }

// https://stripe.com/docs/api/external_account_bank_accounts/retrieve
// external_account_id is prepended with ba_
func (s *Stripe) GetExternalAccountBankAccount(external_account_id string) (*stripe.BankAccount, error) {
	ba, err := s.client.BankAccounts.Get(external_account_id, nil)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/external_account_bank_accounts/update
// external_account_id is prepended with ba_
func (s *Stripe) UpdateExternalAccountBankAccount(external_account_id string, params *stripe.BankAccountParams) (*stripe.BankAccount, error) {
	ba, err := s.client.BankAccounts.Update(external_account_id, params)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/external_account_bank_accounts/delete
// external_account_id is prepended with ba_
func (s *Stripe) DeleteExternalAccountBankAccount(external_account_id string) (*stripe.BankAccount, error) {
	ba, err := s.client.BankAccounts.Del(external_account_id, nil)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

// https://stripe.com/docs/api/external_account_bank_accounts/list
// Iterate through external account bank accounts
/*
for i.Next() {
	  ba := i.BankAccount()
}
*/
func (s *Stripe) ListExternalAccountBankAccounts(opt_params *stripe.BankAccountListParams) *bankaccount.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.BankAccounts.List(opt_params)
	return i
}

// https://stripe.com/docs/api/external_account_cards/create
// Required Params - external_account
func (s *Stripe) CreateExternalAccountCard(params *stripe.CardParams) (*stripe.Card, error) {
	c, err := s.client.Cards.New(params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/external_account_cards/retrieve
// external_account_id is prepended with card_
func (s *Stripe) GetExternalAccountCard(external_account_id string) (*stripe.Card, error) {
	c, err := s.client.Cards.Get(external_account_id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/external_account_cards/update
func (s *Stripe) UpdateExternalAccountCard(external_account_id string, params *stripe.CardParams) (*stripe.Card, error) {
	c, err := s.client.Cards.Update(external_account_id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/external_account_cards/delete
// external_account_id is prepended with card_
func (s *Stripe) DeleteExternalAccountCard(external_account_id string) (*stripe.Card, error) {
	c, err := s.client.Cards.Del(external_account_id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/external_account_cards/list
// Iterate through external account cards
/*
for i.Next() {
	  c := i.Card()
}
*/
func (s *Stripe) ListExternalAccountCards(opt_params *stripe.CardListParams) *card.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Cards.List(opt_params)
	return i
}

//SECTION - PERSON
/*
	This is an object representing a person associated with a Stripe account.

	A platform cannot access a Standard or Express account's persons after the account starts onboarding, such as after generating an account link for the account. See the Standard onboarding or Express onboarding documentation for information about platform prefilling and account onboarding steps.
*/

// https://stripe.com/docs/api/persons/create
func (s *Stripe) CreatePerson(account_id string, params *stripe.PersonParams) (*stripe.Person, error) {
	params.Account = stripe.String(account_id)

	p, err := s.client.Persons.New(params)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/persons/retrieve
// person_id is prepended with person_
func (s *Stripe) GetPerson(person_id string) (*stripe.Person, error) {
	p, err := s.client.Persons.Get(person_id, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/persons/update
// person_id is prepended with person_
func (s *Stripe) UpdatePerson(person_id string, params *stripe.PersonParams) (*stripe.Person, error) {
	p, err := s.client.Persons.Update(person_id, params)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/persons/delete
// person_id is prepended with person_
func (s *Stripe) DeletePerson(person_id string) (*stripe.Person, error) {
	p, err := s.client.Persons.Del(person_id, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// https://stripe.com/docs/api/persons/list

//SECTION - TOP-UPS
/*
To top up your Stripe balance, you create a top-up object. You can retrieve individual top-ups, as well as list all top-ups. Top-ups are identified by a unique, random ID.
*/

// https://stripe.com/docs/api/topups/create
// Required Params - amount, currency
func (s *Stripe) CreateTopUp(amount int64, currency_3ISO string, params *stripe.TopupParams) (*stripe.Topup, error) {
	params.Amount = stripe.Int64(amount)
	params.Currency = stripe.String(currency_3ISO)

	tu, err := s.client.Topups.New(params)
	if err != nil {
		return nil, err
	}

	return tu, nil
}

// https://stripe.com/docs/api/topups/retrieve
// topup_id is prepended with tu_
func (s *Stripe) GetTopUp(topup_id string) (*stripe.Topup, error) {
	tu, err := s.client.Topups.Get(topup_id, nil)
	if err != nil {
		return nil, err
	}

	return tu, nil
}

// https://stripe.com/docs/api/topups/update
// topup_id is prepended with tu_
func (s *Stripe) UpdateTopUp(topup_id string, params *stripe.TopupParams) (*stripe.Topup, error) {
	tu, err := s.client.Topups.Update(topup_id, params)
	if err != nil {
		return nil, err
	}

	return tu, nil
}

// https://stripe.com/docs/api/topups/list
// Iterate through topups
/*
for i.Next() {
	  tu := i.Topup()
}
*/
func (s *Stripe) ListTopUps(opt_params *stripe.TopupListParams) *topup.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Topups.List(opt_params)
	return i
}

// https://stripe.com/docs/api/topups/cancel
// topup_id is prepended with tu_
func (s *Stripe) CancelTopUp(topup_id string) (*stripe.Topup, error) {
	t, err := s.client.Topups.Cancel(topup_id, nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

//SECTION - TRANSFERS
/*
	A Transfer object is created when you move funds between Stripe accounts as part of Connect.
*/

// https://stripe.com/docs/api/transfers/create
// Required Params - amount, currency, destination (account id)
func (s *Stripe) CreateTransfer(amount int64, currency_3ISO, destination_account_id string, params *stripe.TransferParams) (*stripe.Transfer, error) {
	params.Amount = stripe.Int64(amount)
	params.Currency = stripe.String(currency_3ISO)
	params.Destination = stripe.String(destination_account_id)

	t, err := s.client.Transfers.New(params)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/transfers/retrieve
// transfer_id is prepended with tr_
func (s *Stripe) GetTransfer(transfer_id string) (*stripe.Transfer, error) {
	t, err := s.client.Transfers.Get(transfer_id, nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/transfers/update
// transfer_id is prepended with tr_
func (s *Stripe) UpdateTransfer(transfer_id string, params *stripe.TransferParams) (*stripe.Transfer, error) {
	t, err := s.client.Transfers.Update(transfer_id, params)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/transfers/list
// Iterate through transfers
/*
for i.Next() {
	  t := i.Transfer()
}
*/
func (s *Stripe) ListTransfers(opt_params *stripe.TransferListParams) *transfer.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Transfers.List(opt_params)
	return i
}

//SECTION - TRANSFER REVERSALS
/*
	Stripe Connect platforms can reverse transfers made to a connected account, either entirely or partially, and can also specify whether to refund any related application fees. Transfer reversals add to the platform's balance and subtract from the destination account's balance.

	Reversing a transfer that was made for a destination charge is allowed only up to the amount of the charge. It is possible to reverse a transfer_group transfer only if the destination account has enough balance to cover the reversal.
*/

// https://stripe.com/docs/api/transfer_reversals/create
func (s *Stripe) CreateTransferReversal(params *stripe.TransferReversalParams) (*stripe.TransferReversal, error) {
	r, err := s.client.TransferReversals.New(params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/transfer_reversals/retrieve
// transfer_reversal_id is prepended with trr_
func (s *Stripe) GetTransferReversal(transfer_reversal_id string) (*stripe.TransferReversal, error) {
	r, err := s.client.TransferReversals.Get(transfer_reversal_id, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/transfer_reversals/update
// transfer_reversal_id is prepended with trr_
func (s *Stripe) UpdateTransferReversal(transfer_reversal_id string, params *stripe.TransferReversalParams) (*stripe.TransferReversal, error) {
	r, err := s.client.TransferReversals.Update(transfer_reversal_id, params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/transfer_reversals/list
// Iterate through transfer reversals
/*
for i.Next() {
	  r := i.TransferReversal() -- TODO: Double check this
}
*/
func (s *Stripe) ListTransferReversals(opt_params *stripe.TransferReversalListParams) *transferreversal.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.TransferReversals.List(opt_params)
	return i
}

//SECTION - SECRETS
/*
Secret Store is an API that allows Stripe Apps developers to securely persist secrets for use by UI Extensions and app backends.

	The primary resource in Secret Store is a secret. Other apps can't view secrets created by an app. Additionally, secrets are scoped to provide further permission control.

	All Dashboard users and the app backend share account scoped secrets. Use the account scope for secrets that don't change per-user, like a third-party API key.

	A user scoped secret is accessible by the app backend and one specific Dashboard user. Use the user scope for per-user secrets like per-user OAuth tokens, where different users might have different permissions.
*/

// https://stripe.com/docs/api/apps/secret_store/set
// Required Params - name, payload, scope
func (s *Stripe) SetSecret(name, payload, scope string, user_id *string, params *stripe.AppsSecretParams) (*stripe.AppsSecret, error) {
	params.Name = stripe.String(name)
	params.Payload = stripe.String(payload)
	params.Scope = &stripe.AppsSecretScopeParams{
		Type: stripe.String(scope),
	}

	if user_id != nil {
		params.Scope.User = stripe.String(*user_id)
	}

	ss, err := s.client.AppsSecrets.New(params)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// https://stripe.com/docs/api/apps/secret_store/find
// Required Params - name, scope
func (s *Stripe) GetSecret(name, scope string, user_id *string, params *stripe.AppsSecretFindParams) (*stripe.AppsSecret, error) {
	params.Name = stripe.String(name)
	params.Scope = &stripe.AppsSecretFindScopeParams{
		Type: stripe.String(scope),
	}

	if user_id != nil {
		params.Scope.User = stripe.String(*user_id)
	}

	ss, err := s.client.AppsSecrets.Find(params)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// https://stripe.com/docs/api/apps/secret_store/delete
// Required Params - name, scope
func (s *Stripe) DeleteSecret(name, scope string, user_id *string, params *stripe.AppsSecretDeleteWhereParams) (*stripe.AppsSecret, error) {
	params.Name = stripe.String(name)
	params.Scope = &stripe.AppsSecretDeleteWhereScopeParams{
		Type: stripe.String(scope),
	}

	if user_id != nil {
		params.Scope.User = stripe.String(*user_id)
	}

	ss, err := s.client.AppsSecrets.DeleteWhere(params)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// https://stripe.com/docs/api/apps/secret_store/list
// Iterate through secrets
/*
for i.Next() {
	  ss := i.AppsSecret()
}
*/
func (s *Stripe) ListSecrets(opt_params *stripe.AppsSecretListParams) *secret.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.AppsSecrets.List(opt_params)
	return i
}

// -- FRAUD -- //
//SECTION - EARLY FRAUD WARNINGS
/*
	An early fraud warning indicates that the card issuer has notified us that a charge may be fraudulent.
*/

// https://stripe.com/docs/api/radar/early_fraud_warnings/retrieve
// early_fraud_warning_id is prepended with issfr_
func (s *Stripe) GetEarlyFraudWarning(early_fraud_warning_id string) (*stripe.RadarEarlyFraudWarning, error) {
	efw, err := s.client.RadarEarlyFraudWarnings.Get(early_fraud_warning_id, nil)
	if err != nil {
		return nil, err
	}

	return efw, nil
}

// https://stripe.com/docs/api/radar/early_fraud_warnings/list
// Iterate through early fraud warnings
/*
for i.Next() {
	  efw := i.RadarEarlyFraudWarning()
}
*/
func (s *Stripe) ListEarlyFraudWarnings(opt_params *stripe.RadarEarlyFraudWarningListParams) *earlyfraudwarning.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.RadarEarlyFraudWarnings.List(opt_params)
	return i
}

//SECTION - REVIEWS
/*
	Reviews can be used to supplement automated fraud detection with human expertise.
*/

// https://stripe.com/docs/api/radar/reviews/approve
// review_id is prepended with prv_
func (s *Stripe) ApproveReview(review_id string) (*stripe.Review, error) {
	r, err := s.client.Reviews.Approve(review_id, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/radar/reviews/retrieve
// review_id is prepended with prv_
func (s *Stripe) GetReview(review_id string) (*stripe.Review, error) {
	r, err := s.client.Reviews.Get(review_id, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/radar/reviews/list
// Iterate through reviews
/*
for i.Next() {
	  r := i.Review()
}
*/
func (s *Stripe) ListReviews(opt_params *stripe.ReviewListParams) *review.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.Reviews.List(opt_params)
	return i
}

//SECTION - VALUE LISTS
/*
	Value lists allow you to group values together which can then be referenced in rules.
*/

// https://stripe.com/docs/api/radar/value_lists/create
// Required Params - alias, name
func (s *Stripe) CreateValueList(alias, name string, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	params.Alias = stripe.String(alias)
	params.Name = stripe.String(name)

	vl, err := s.client.RadarValueLists.New(params)
	if err != nil {
		return nil, err
	}

	return vl, nil
}

// https://stripe.com/docs/api/radar/value_lists/retrieve
// value_list_id is prepended with rsl_
func (s *Stripe) GetValueList(value_list_id string) (*stripe.RadarValueList, error) {
	vl, err := s.client.RadarValueLists.Get(value_list_id, nil)
	if err != nil {
		return nil, err
	}

	return vl, nil
}

// https://stripe.com/docs/api/radar/value_lists/update
// value_list_id is prepended with rsl_
func (s *Stripe) UpdateValueList(value_list_id string, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	vl, err := s.client.RadarValueLists.Update(value_list_id, params)
	if err != nil {
		return nil, err
	}

	return vl, nil
}

// https://stripe.com/docs/api/radar/value_lists/delete
// value_list_id is prepended with rsl_
func (s *Stripe) DeleteValueList(value_list_id string) (*stripe.RadarValueList, error) {
	vl, err := s.client.RadarValueLists.Del(value_list_id, nil)
	if err != nil {
		return nil, err
	}

	return vl, nil
}

// https://stripe.com/docs/api/radar/value_lists/list
// Iterate through value lists
/*
for i.Next() {
	  vl := i.RadarValueList()
}
*/
func (s *Stripe) ListValueLists(opt_params *stripe.RadarValueListListParams) *valuelist.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.RadarValueLists.List(opt_params)
	return i
}

//SECTION - VALUE LIST ITEMS
/*
	Value list items allow you to add specific values to a given Radar value list, which can then be used in rules.
*/

// https://stripe.com/docs/api/radar/value_list_items/create
// Required Params - value, value_list
func (s *Stripe) CreateValueListItem(value, value_list_id string, params *stripe.RadarValueListItemParams) (*stripe.RadarValueListItem, error) {
	params.Value = stripe.String(value)
	params.ValueList = stripe.String(value_list_id)

	vli, err := s.client.RadarValueListItems.New(params)
	if err != nil {
		return nil, err
	}

	return vli, nil
}

// https://stripe.com/docs/api/radar/value_list_items/retrieve
// value_list_item_id is prepended with rsli_
func (s *Stripe) GetValueListItem(value_list_item_id string) (*stripe.RadarValueListItem, error) {
	vli, err := s.client.RadarValueListItems.Get(value_list_item_id, nil)
	if err != nil {
		return nil, err
	}

	return vli, nil
}

// https://stripe.com/docs/api/radar/value_list_items/delete
// value_list_item_id is prepended with rsli_
func (s *Stripe) DeleteValueListItem(value_list_item_id string) (*stripe.RadarValueListItem, error) {
	vli, err := s.client.RadarValueListItems.Del(value_list_item_id, nil)
	if err != nil {
		return nil, err
	}

	return vli, nil
}

// https://stripe.com/docs/api/radar/value_list_items/list
// Iterate through value list items
/*
for i.Next() {
	  vli := i.RadarValueListItem()
}
*/
func (s *Stripe) ListValueListItems(opt_params *stripe.RadarValueListItemListParams) *valuelistitem.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.RadarValueListItems.List(opt_params)
	return i
}

// -- ISSUING -- //
//SECTION - AUTHORIZATIONS
/*
	When an issued card is used to make a purchase, an Issuing Authorization object is created. Authorizations must be approved for the purchase to be completed successfully.
*/

// https://stripe.com/docs/api/issuing/authorizations/retrieve
// authorization_id is prepended with iauth_
func (s *Stripe) GetAuthorization(authorization_id string) (*stripe.IssuingAuthorization, error) {
	a, err := s.client.IssuingAuthorizations.Get(authorization_id, nil)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/issuing/authorizations/update
// authorization_id is prepended with iauth_
func (s *Stripe) UpdateAuthorization(authorization_id string, params *stripe.IssuingAuthorizationParams) (*stripe.IssuingAuthorization, error) {
	a, err := s.client.IssuingAuthorizations.Update(authorization_id, params)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/issuing/authorizations/approve
// authorization_id is prepended with iauth_
func (s *Stripe) ApproveAuthorization(authorization_id string) (*stripe.IssuingAuthorization, error) {
	a, err := s.client.IssuingAuthorizations.Approve(authorization_id, nil)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/issuing/authorizations/decline
// authorization_id is prepended with iauth_
func (s *Stripe) DeclineAuthorization(authorization_id string) (*stripe.IssuingAuthorization, error) {
	a, err := s.client.IssuingAuthorizations.Decline(authorization_id, nil)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// https://stripe.com/docs/api/issuing/authorizations/list
// Iterate through authorizations
/*
for i.Next() {
	  a := i.IssuingAuthorization()
}
*/
func (s *Stripe) ListAuthorizations(opt_params *stripe.IssuingAuthorizationListParams) *authorization.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.IssuingAuthorizations.List(opt_params)
	return i
}

//SECTION - CARDHOLDERS
/*
	An Issuing Cardholder object represents an individual or business entity who is issued cards.
*/

// https://stripe.com/docs/api/issuing/cardholders/create
// Required Params - name, billing (cardholders billing address)
func (s *Stripe) CreateCardholder(name string, billing *stripe.IssuingCardholderBillingParams, params *stripe.IssuingCardholderParams) (*stripe.IssuingCardholder, error) {
	params.Name = stripe.String(name)
	params.Billing = billing

	ch, err := s.client.IssuingCardholders.New(params)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

// https://stripe.com/docs/api/issuing/cardholders/retrieve
// cardholder_id is prepended with ich_
func (s *Stripe) GetCardholder(cardholder_id string) (*stripe.IssuingCardholder, error) {
	ch, err := s.client.IssuingCardholders.Get(cardholder_id, nil)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

// https://stripe.com/docs/api/issuing/cardholders/update
// cardholder_id is prepended with ich_
func (s *Stripe) UpdateCardholder(cardholder_id string, params *stripe.IssuingCardholderParams) (*stripe.IssuingCardholder, error) {
	ch, err := s.client.IssuingCardholders.Update(cardholder_id, params)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

// https://stripe.com/docs/api/issuing/cardholders/list
// Iterate through cardholders
/*
for i.Next() {
	  ch := i.IssuingCardholder()
}
*/
func (s *Stripe) ListCardholders(opt_params *stripe.IssuingCardholderListParams) *cardholder.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.IssuingCardholders.List(opt_params)
	return i
}

//SECTION - CARDS
/*
	You can create physical or virtual cards that are issued to cardholders.
*/

// https://stripe.com/docs/api/issuing/cards/create
// Required Params - cardholder, currency, card_type (physical or virtual)
func (s *Stripe) CreateIssuingCard(cardholder_id, currency_3ISO, card_type string, params *stripe.IssuingCardParams) (*stripe.IssuingCard, error) {
	params.Cardholder = stripe.String(cardholder_id)
	params.Currency = stripe.String(currency_3ISO)
	params.Type = stripe.String(card_type)

	c, err := s.client.IssuingCards.New(params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/issuing/cards/retrieve
// card_id is prepended with ic_
func (s *Stripe) GetIssuingCard(card_id string) (*stripe.IssuingCard, error) {
	c, err := s.client.IssuingCards.Get(card_id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/issuing/cards/update
func (s *Stripe) UpdateIssuingCard(card_id string, params *stripe.IssuingCardParams) (*stripe.IssuingCard, error) {
	c, err := s.client.IssuingCards.Update(card_id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/issuing/cards/list
// Iterate through cards
/*
for i.Next() {
	  c := i.IssuingCard()
}
*/
func (s *Stripe) ListIssuingCards(opt_params *stripe.IssuingCardListParams) *issuingCard.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.IssuingCards.List(opt_params)
	return i
}

// https://stripe.com/docs/api/issuing/cards/test_mode_ship
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/issuing/cards/test_mode_deliver
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/issuing/cards/test_mode_return
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/issuing/cards/test_mode_fail
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/issuing/cards/test_mode_submit
// TODO - stripe client doesn't implement this - need to do with http client

//SECTION - DISPUTES
/*
	As a card issuer, you can dispute transactions that the cardholder does not recognize, suspects to be fraudulent, or has other issues with.
*/

// https://stripe.com/docs/api/issuing/disputes/create
func (s *Stripe) CreateIssuingDispute(params *stripe.IssuingDisputeParams) (*stripe.IssuingDispute, error) {
	d, err := s.client.IssuingDisputes.New(params)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// https://stripe.com/docs/api/issuing/dispute/submit
// dispute_id is prepended with idp_
func (s *Stripe) SubmitIssuingDispute(dispute_id string) (*stripe.IssuingDispute, error) {
	d, err := s.client.IssuingDisputes.Submit(dispute_id, nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// https://stripe.com/docs/api/issuing/disputes/retrieve
// dispute_id is prepended with idp_
func (s *Stripe) GetIssuingDispute(dispute_id string) (*stripe.IssuingDispute, error) {
	d, err := s.client.IssuingDisputes.Get(dispute_id, nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// https://stripe.com/docs/api/issuing/disputes/update
// dispute_id is prepended with idp_
func (s *Stripe) UpdateIssuingDispute(dispute_id string, params *stripe.IssuingDisputeParams) (*stripe.IssuingDispute, error) {
	d, err := s.client.IssuingDisputes.Update(dispute_id, params)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// https://stripe.com/docs/api/issuing/disputes/list
// Iterate through disputes
/*
for i.Next() {
	  d := i.IssuingDispute()
}
*/
func (s *Stripe) ListIssuingDisputes(opt_params *stripe.IssuingDisputeListParams) *issuingDispute.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.IssuingDisputes.List(opt_params)
	return i
}

//SECTION - FUNDING INSTRUCTIONS
/*
	Funding Instructions contain reusable bank account and routing information. Push funds to these addresses via bank transfer to top up Issuing Balances.
*/

// https://stripe.com/docs/api/issuing/funding_instructions/create
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/issuing/funding_instructions/list
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/issuing/funding_instructions/fund
// TODO - stripe client doesn't implement this - need to do with http client

//SECTION - TRANSACTIONS
/*
	Any use of an issued card that results in funds entering or leaving your Stripe account, such as a completed purchase or refund, is represented by an Issuing Transaction object.
*/

// https://stripe.com/docs/api/issuing/transactions/retrieve
// transaction_id is prepended with ipi_
func (s *Stripe) GetIssuingTransaction(transaction_id string) (*stripe.IssuingTransaction, error) {
	t, err := s.client.IssuingTransactions.Get(transaction_id, nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/issuing/transactions/update
// transaction_id is prepended with ipi_
func (s *Stripe) UpdateIssuingTransaction(transaction_id string, params *stripe.IssuingTransactionParams) (*stripe.IssuingTransaction, error) {
	t, err := s.client.IssuingTransactions.Update(transaction_id, params)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// https://stripe.com/docs/api/issuing/transactions/list
// Iterate through transactions
/*
for i.Next() {
	  t := i.IssuingTransaction()
}
*/
func (s *Stripe) ListIssuingTransactions(opt_params *stripe.IssuingTransactionListParams) *transaction.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.IssuingTransactions.List(opt_params)
	return i
}

// -- TERMINAL -- //
//SECTION - CONNECTION TOKEN
/*
	A Connection Token is used by the Stripe Terminal SDK to connect to a reader.
*/

// https://stripe.com/docs/api/terminal/connection_tokens/create
func (s *Stripe) CreateTerminalConnectionToken(params *stripe.TerminalConnectionTokenParams) (*stripe.TerminalConnectionToken, error) {
	ct, err := s.client.TerminalConnectionTokens.New(params)
	if err != nil {
		return nil, err
	}

	return ct, nil
}

//SECTION - LOCATION
/*
	A Location represents a grouping of readers.
*/

// https://stripe.com/docs/api/terminal/locations/create
// Required Params - display_name, address
func (s *Stripe) CreateTerminalLocation(display_name string, address *stripe.AddressParams, params *stripe.TerminalLocationParams) (*stripe.TerminalLocation, error) {
	params.DisplayName = stripe.String(display_name)
	params.Address = address

	l, err := s.client.TerminalLocations.New(params)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// https://stripe.com/docs/api/terminal/locations/retrieve
// location_id is prepended with tml_
func (s *Stripe) GetTerminalLocation(location_id string) (*stripe.TerminalLocation, error) {
	l, err := s.client.TerminalLocations.Get(location_id, nil)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// https://stripe.com/docs/api/terminal/locations/update
// location_id is prepended with tml_
func (s *Stripe) UpdateTerminalLocation(location_id string, params *stripe.TerminalLocationParams) (*stripe.TerminalLocation, error) {
	l, err := s.client.TerminalLocations.Update(location_id, params)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// https://stripe.com/docs/api/terminal/locations/delete
// location_id is prepended with tml_
func (s *Stripe) DeleteTerminalLocation(location_id string) (*stripe.TerminalLocation, error) {
	l, err := s.client.TerminalLocations.Del(location_id, nil)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// https://stripe.com/docs/api/terminal/locations/list
// Iterate through locations
/*
for i.Next() {
	  l := i.TerminalLocation()
}
*/
func (s *Stripe) ListTerminalLocations(opt_params *stripe.TerminalLocationListParams) *location.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.TerminalLocations.List(opt_params)
	return i
}

//SECTION - READER
/*
	A Reader represents a physical device for accepting payment details.
*/

// https://stripe.com/docs/api/terminal/readers/create
// Required Params - registration_code, location
func (s *Stripe) CreateTerminalReader(registration_code, location_id string, params *stripe.TerminalReaderParams) (*stripe.TerminalReader, error) {
	params.RegistrationCode = stripe.String(registration_code)
	params.Location = stripe.String(location_id)

	r, err := s.client.TerminalReaders.New(params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/terminal/readers/retrieve
// reader_id is prepended with tmr_
func (s *Stripe) GetTerminalReader(reader_id string) (*stripe.TerminalReader, error) {
	r, err := s.client.TerminalReaders.Get(reader_id, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/terminal/readers/update
// reader_id is prepended with tmr_
func (s *Stripe) UpdateTerminalReader(reader_id string, params *stripe.TerminalReaderParams) (*stripe.TerminalReader, error) {
	r, err := s.client.TerminalReaders.Update(reader_id, params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/terminal/readers/delete
// reader_id is prepended with tmr_
func (s *Stripe) DeleteTerminalReader(reader_id string) (*stripe.TerminalReader, error) {
	r, err := s.client.TerminalReaders.Del(reader_id, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/terminal/readers/list
// Iterate through readers
/*
for i.Next() {
	  r := i.TerminalReader()
}
*/
func (s *Stripe) ListTerminalReaders(opt_params *stripe.TerminalReaderListParams) *reader.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.TerminalReaders.List(opt_params)
	return i
}

// https://stripe.com/docs/api/terminal/readers/process_payment_intent
// Required Params - payment_intent, reader_id
func (s *Stripe) ProcessTerminalPaymentIntent(payment_intent_id, reader_id string) (*stripe.TerminalReader, error) {
	params := &stripe.TerminalReaderProcessPaymentIntentParams{
		PaymentIntent: stripe.String(payment_intent_id),
	}

	r, err := s.client.TerminalReaders.ProcessPaymentIntent(reader_id, params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/terminal/readers/process_setup_intent
// Required Params - customer_consent_collected, setup_intent, reader_id
func (s *Stripe) ProcessTerminalSetupIntent(customer_consent_collected bool, setup_intent_id, reader_id string) (*stripe.TerminalReader, error) {
	params := &stripe.TerminalReaderProcessSetupIntentParams{
		CustomerConsentCollected: stripe.Bool(customer_consent_collected),
		SetupIntent:              stripe.String(setup_intent_id),
	}

	r, err := s.client.TerminalReaders.ProcessSetupIntent(reader_id, params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/terminal/readers/set_reader_display
// Required Params - type (only 'cart') - 12/8/20 -
func (s *Stripe) SetTerminalReaderDisplay(reader_id, type_ string, params *stripe.TerminalReaderSetReaderDisplayParams) (*stripe.TerminalReader, error) {
	params.Type = stripe.String(type_)

	r, err := s.client.TerminalReaders.SetReaderDisplay(reader_id, params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/terminal/readers/refund_payment
func (s *Stripe) RefundChargeOrPaymentIntentInPerson(reader_id, payment_intent_id string) (*stripe.TerminalReader, error) {
	params := &stripe.TerminalReaderRefundPaymentParams{
		PaymentIntent: stripe.String(payment_intent_id),
	}

	r, err := s.client.TerminalReaders.RefundPayment(reader_id, params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// https://stripe.com/docs/api/terminal/readers/cancel_action
func (s *Stripe) CancelTerminalReaderAction(reader_id string) (*stripe.TerminalReader, error) {
	r, err := s.client.TerminalReaders.CancelAction(reader_id, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

//SECTION - TERMINAL HARDWARE ORDER
/*
	A TerminalHardwareOrder represents an order for Terminal hardware, containing information such as the price, shipping information and the items ordered.
*/

// https://stripe.com/docs/api/terminal/hardware_orders/preview
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/create
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/retrieve
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/list
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/confirm
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/cancel
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/test_mode_mark_ready_to_ship
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/test_mode_ship
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/test_mode_deliver
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_orders/test_mode_mark_undeliverable
// TODO - stripe client doesn't implement this - need to do with http client

//SECTION - TERMINAL HARDWARE PRODUCT
/*
	A TerminalHardwareProduct is a category of hardware devices that are generally similar, but may have variations depending on the country it’s shipped to.

	TerminalHardwareSKUs represent variations within the same Product (for example, a country specific device). For example, WisePOS E is a TerminalHardwareProduct and a WisePOS E - US and WisePOS E - UK are TerminalHardwareSKUs.
*/

// https://stripe.com/docs/api/terminal/hardware_products/retrieve
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_products/list
// TODO - stripe client doesn't implement this - need to do with http client

//SECTION - TERMINAL HARDWARE SKU
/*
	A TerminalHardwareSKU represents a SKU for Terminal hardware. A SKU is a representation of a product available for purchase, containing information such as the name, price, and images.
*/

// https://stripe.com/docs/api/terminal/hardware_skus/retrieve
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_skus/list
// TODO - stripe client doesn't implement this - need to do with http client

//SECTION - TERMINAL HARDWARE SHIPPING METHOD
/*
	A TerminalHardwareShipping represents a Shipping Method for Terminal hardware. A Shipping Method is a country-specific representation of a way to ship hardware, containing information such as the country, name, and expected delivery date.
*/

// https://stripe.com/docs/api/terminal/hardware_shipping_methods/retrieve
// TODO - stripe client doesn't implement this - need to do with http client

// https://stripe.com/docs/api/terminal/hardware_shipping_methods/list
// TODO - stripe client doesn't implement this - need to do with http client

//SECTION - CONFIGURATIONS
/*
	A Configurations object represents how features should be configured for terminal readers.
*/

// https://stripe.com/docs/api/terminal/configuration/create
func (s *Stripe) CreateTerminalConfiguration(params *stripe.TerminalConfigurationParams) (*stripe.TerminalConfiguration, error) {
	c, err := s.client.TerminalConfigurations.New(params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/terminal/configuration/retrieve
// configuration_id is prepended with tmc_
func (s *Stripe) GetTerminalConfiguration(configuration_id string) (*stripe.TerminalConfiguration, error) {
	c, err := s.client.TerminalConfigurations.Get(configuration_id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/terminal/configuration/update
// configuration_id is prepended with tmc_
func (s *Stripe) UpdateTerminalConfiguration(configuration_id string, params *stripe.TerminalConfigurationParams) (*stripe.TerminalConfiguration, error) {
	c, err := s.client.TerminalConfigurations.Update(configuration_id, params)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/terminal/configuration/delete
// configuration_id is prepended with tmc_
func (s *Stripe) DeleteTerminalConfiguration(configuration_id string) (*stripe.TerminalConfiguration, error) {
	c, err := s.client.TerminalConfigurations.Del(configuration_id, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// https://stripe.com/docs/api/terminal/configuration/list
// Iterate through configurations
/*
for i.Next() {
	  c := i.TerminalConfiguration()
}
*/
func (s *Stripe) ListTerminalConfigurations(opt_params *stripe.TerminalConfigurationListParams) *configuration.Iter {
	opt_params.Filters.AddFilter("limit", "", "3")
	i := s.client.TerminalConfigurations.List(opt_params)
	return i
}

// -- TREASURY -- //
//SECTION - FINANCIAL ACCOUNTS

//SECTION - FINANCIAL ACCOUNT FEATURES

//SECTION - FINANCIAL ACCOUNT TRANSACTIONS

//SECTION - TRANSACTION ENTRIES

//SECTION - OUTBOUND TRANSFERS

//SECTION - OUTBOUND PAYMENTS

//SECTION - INBOUND TRANSFERS

//SECTION - RECEIVED CREDITS

//SECTION - RECEIVED DEBITS

//SECTION - CREDIT REVERSALS

//SECTION - DEBIT REVERSALS

// -- SIGMA -- //
//SECTION - SCHEDULED QUERIES

// -- REPORTING -- //
//SECTION - REPORT RUNS

//SECTION - REPORT TYPES

// -- FINANCIAL CONNECTIONS -- //
//SECTION - ACCOUNTS

//SECTION - INFERRED ACCOUNT BALANCE

//SECTION - ACCOUNT OWNERSHIP

//SECTION - SESSION

//SECTION - TRANSACTIONS

// -- TAX -- //
//SECTION - TAX SETTINGS

//SECTION - TAX CALCULATIONS

//SECTION - TAX TRANSACTIONS

// -- IDENTITY -- //
//SECTION - VERIFICATION SESSION

//SECTION - VERIFICATION REPORT

// -- CRYPTO -- //
//leaving this out for now. As Stripe may not be the best choice for this.

// -- WEBHOOKS -- //
//SECTION - WEBHOOK ENDPOINTS

package gofabric

import (
	"database/sql"
	"os"
	"time"

	"github.com/lordwestcott/gofabric/messaging"
	"github.com/lordwestcott/gofabric/openai"
	"github.com/lordwestcott/gofabric/stripe"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

type App struct {
	// Models    data.Models //Must be implemented on project using this package.
	DB        *sql.DB
	Stripe    *stripe.Stripe
	Messaging *messaging.Messaging
	OpenAI    *openai.OpenAI
}

func InitApp(envFile string, dsn string) (*App, error) {
	app := &App{}

	if err := godotenv.Load(envFile); err != nil {
		color.Yellow("No .env file found")
		return nil, err
	}

	if os.Getenv("DATABASE_URL") != "" {

		db, err := OpenDB(0, os.Getenv("DATABASE_URL"))
		if err != nil {
			return nil, err
		}
		app.DB = db
	}

	if os.Getenv("STRIPE_PRIVATE_KEY") != "" {
		stripe := stripe.NewStripe(os.Getenv("STRIPE_PRIVATE_KEY"))
		app.Stripe = stripe
	}

	if os.Getenv("TWILIO_AUTH_TOKEN") != "" {
		messaging, err := messaging.NewMessaging()
		if err != nil {
			return nil, err
		}
		app.Messaging = messaging
	}

	if os.Getenv("OPENAI_API_KEY") != "" {
		openAI, err := openai.NewOpenAI()
		if err != nil {
			return nil, err
		}

		app.OpenAI = openAI
	}

	return app, nil
}

func OpenDB(attempt int, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		if attempt > 5 {
			color.Red("Error opening database: %v", err)
			return nil, err
		}
		color.Yellow("Error opening database: %v", err)
		color.Yellow("Backing off for 5 seconds...")
		<-time.After(5 * time.Second)
		attempt++
		return OpenDB(attempt, dsn)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

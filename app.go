package gofabric

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/lordwestcott/gofabric/messaging"
	"github.com/lordwestcott/gofabric/openai"
	"github.com/lordwestcott/gofabric/signin/google"
	"github.com/lordwestcott/gofabric/stripe"
	"github.com/lordwestcott/gofabric/urlsigner"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

type App struct {
	// Models    data.Models //Must be implemented on project using this package.
	DB           *sql.DB
	Stripe       *stripe.Stripe
	Messaging    *messaging.Messaging
	OpenAI       *openai.OpenAI
	URLSigner    *urlsigner.Signer
	ErrorLog     *log.Logger
	InfoLog      *log.Logger
	GoogleSignIn *google.GoogleSignIn
}

func InitApp(envFile string) (*App, error) {
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

	if os.Getenv("GOOGLE_CLIENT_ID") != "" {
		g := google.GoogleSignIn{}
		g.New(os.Getenv("GOOGLE_CLIENT_ID"))

		app.GoogleSignIn = &g
	}

	infoLog, errorLog := app.startLoggers()

	app.ErrorLog = errorLog
	app.InfoLog = infoLog

	if os.Getenv("SIGNING_KEY") != "" {
		signer := urlsigner.Signer{}
		signer.New(os.Getenv("SIGNING_KEY"))
		app.URLSigner = &signer
	} else {
		app.URLSigner = nil
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

func (a *App) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

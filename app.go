package gofabric

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/lordwestcott/gofabric/helpers"
	"github.com/lordwestcott/gofabric/jwt"
	"github.com/lordwestcott/gofabric/messaging"
	"github.com/lordwestcott/gofabric/openai"
	"github.com/lordwestcott/gofabric/session"
	"github.com/lordwestcott/gofabric/signin/oauth"
	"github.com/lordwestcott/gofabric/stripe"
	"github.com/lordwestcott/gofabric/urlsigner"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/fatih/color"
)

type App struct {
	// Models    data.Models //Must be implemented on project using this package.
	DB            *sql.DB
	Stripe        *stripe.Stripe
	Messaging     *messaging.Messaging
	OpenAI        *openai.OpenAI
	URLSigner     *urlsigner.Signer
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	Google_OAuth2 *oauth.Google_OAuth2
	Helpers       *helpers.Helpers
	Session       *scs.SessionManager
	JWT           *jwt.JWT
	Host          string
}

func InitApp() (*App, error) {
	app := &App{}

	app.Host = os.Getenv("HOST")

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

	if os.Getenv("GOOGLE_CLIENT_ID") != "" &&
		os.Getenv("GOOGLE_CLIENT_SECRET") != "" &&
		os.Getenv("GOOGLE_STATE") != "" &&
		os.Getenv("GOOGLE_REDIRECT") != "" {
		g := oauth.Google_OAuth2{}
		redirect := app.Host + os.Getenv("GOOGLE_REDIRECT")
		err := g.New(redirect, os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_STATE"))
		if err != nil {
			return nil, err
		}

		app.Google_OAuth2 = &g
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

	app.Helpers = &helpers.Helpers{}

	if os.Getenv("COOKIE_LIFETIME") != "" &&
		os.Getenv("COOKIE_PERSISTS") != "" &&
		os.Getenv("COOKIE_NAME") != "" &&
		os.Getenv("COOKIE_DOMAIN") != "" &&
		os.Getenv("SESSION_TYPE") != "" &&
		os.Getenv("COOKIE_SECURE") != "" {
		ses := session.Session{
			CookieLifetime: os.Getenv("COOKIE_LIFETIME"),
			CookiePersist:  os.Getenv("COOKIE_PERSISTS"),
			CookieName:     os.Getenv("COOKIE_NAME"),
			CookieDomain:   os.Getenv("COOKIE_DOMAIN"),
			SessionType:    os.Getenv("SESSION_TYPE"),
			CookieSecure:   os.Getenv("COOKIE_SECURE"),
		}
		ses.DBPool = app.DB
		app.Session = ses.InitSession()
	}

	if os.Getenv("JWT_SECRET") != "" {
		app.JWT = &jwt.JWT{
			Secret: []byte(os.Getenv("JWT_SECRET")),
		}
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

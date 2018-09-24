package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"fmt"

	"github.com/gorilla/mux"
  "github.com/nicksnyder/go-i18n/i18n"

	common "github.com/tidepool-org/go-common"
	"github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/clients/hakken"
	"github.com/tidepool-org/go-common/clients/highwater"
	"github.com/tidepool-org/go-common/clients/mongo"
	"github.com/tidepool-org/go-common/clients/shoreline"
	"github.com/tidepool-org/hydrophone/api"
	sc "github.com/tidepool-org/hydrophone/clients"
	"github.com/tidepool-org/hydrophone/templates"
	"github.com/tidepool-org/hydrophone/templates_diabeloop"

)

type (
	Config struct {
		clients.Config
		Service disc.ServiceListing  `json:"service"`
		Mongo   mongo.Config         `json:"mongo"`
		Api     api.Config           `json:"hydrophone"`
		Mail    sc.SesNotifierConfig `json:"sesEmail"`
	}
)

func main() {
	var config Config

	if err := common.LoadEnvironmentConfig([]string{"TIDEPOOL_HYDROPHONE_ENV", "TIDEPOOL_HYDROPHONE_SERVICE"}, &config); err != nil {
		log.Panic("Problem loading config ", err)
	}
	var templateEnv = os.Getenv("TIDEPOOL_TEMPLATE")
	log.Printf(templateEnv)
	/*
	 * Hakken setup
	 */
	hakkenClient := hakken.NewHakkenBuilder().
		WithConfig(&config.HakkenConfig).
		Build()

	if err := hakkenClient.Start(); err != nil {
		log.Fatal(err)
	}
	defer hakkenClient.Close()

	/*
	 * Clients
	 */

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient := &http.Client{Transport: tr}

	shoreline := shoreline.NewShorelineClientBuilder().
		WithHostGetter(config.ShorelineConfig.ToHostGetter(hakkenClient)).
		WithHttpClient(httpClient).
		WithConfig(&config.ShorelineConfig.ShorelineClientConfig).
		Build()

	if err := shoreline.Start(); err != nil {
		log.Fatal(err)
	}

	gatekeeper := clients.NewGatekeeperClientBuilder().
		WithHostGetter(config.GatekeeperConfig.ToHostGetter(hakkenClient)).
		WithHttpClient(httpClient).
		WithTokenProvider(shoreline).
		Build()

	highwater := highwater.NewHighwaterClientBuilder().
		WithHostGetter(config.HighwaterConfig.ToHostGetter(hakkenClient)).
		WithHttpClient(httpClient).
		WithConfig(&config.HighwaterConfig.HighwaterClientConfig).
		Build()

	seagull := clients.NewSeagullClientBuilder().
		WithHostGetter(config.SeagullConfig.ToHostGetter(hakkenClient)).
		WithHttpClient(httpClient).
		Build()

	/*
	 * hydrophone setup
	 */
	store := sc.NewMongoStoreClient(&config.Mongo)
	mail := sc.NewSesNotifier(&config.Mail)

	emailTemplates, err := templates.New()
	if templateEnv == "diabeloop" {
		emailTemplates, err = templates_diabeloop.NewDiabeloop()
	}
	if err != nil {
		log.Fatal(err)
	}

	rtr := mux.NewRouter()
	api := api.InitApi(config.Api, store, mail, shoreline, gatekeeper, highwater, seagull, emailTemplates)
	api.SetHandlers("", rtr)

	// Initialisation du type Profile
	type(
		Profile struct {
			Language string  `json:"language"`
		}
	)
	// La variable profile prend la structure de Profile
	var profile = &Profile{};

	// Recherche le profil depuis un id (973205e169) et rempli la variable profile
	seagull.GetCollection("973205e169","profile", shoreline.TokenProvide() ,profile)

	// Initialisation en choisissant la traduction
	i18n.MustLoadTranslationFile("locales/"+profile.Language+".json")
	T, _ := i18n.Tfunc(profile.Language)

	fmt.Printf("%+v\n",T("test"))

	var content map[string]string
	content = make(map[string]string)
	content["WebURL"] = "/test/"
	content["AssetURL"] = "/testAsset/"

	subject, body, err := emailTemplates["password_reset"].Execute(content)
	fmt.Printf("%+v\n",subject)
	fmt.Printf("%+v\n",body)


	/*
	 * Serve it up and publish
	 */
	done := make(chan bool)
	server := common.NewServer(&http.Server{
		Addr:    config.Service.GetPort(),
		Handler: rtr,
	})

	var start func() error
	if config.Service.Scheme == "https" {
		sslSpec := config.Service.GetSSLSpec()
		start = func() error { return server.ListenAndServeTLS(sslSpec.CertFile, sslSpec.KeyFile) }
	} else {
		start = func() error { return server.ListenAndServe() }
	}
	if err := start(); err != nil {
		log.Fatal(err)
	}

	hakkenClient.Publish(&config.Service)

	signals := make(chan os.Signal, 40)
	signal.Notify(signals)
	go func() {
		for {
			sig := <-signals
			log.Printf("Got signal [%s]", sig)

			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				server.Close()
				done <- true
			}
		}
	}()

	<-done

}

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	yaml "gopkg.in/yaml.v2"

	commonClients "github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/clients/highwater"
	"github.com/tidepool-org/go-common/clients/shoreline"
	"github.com/tidepool-org/go-common/clients/status"
	"github.com/tidepool-org/hydrophone/clients"
	"github.com/tidepool-org/hydrophone/models"
)

type (
	Api struct {
		Store          clients.StoreClient
		notifier       clients.Notifier
		templates      models.Templates
		sl             shoreline.Client
		gatekeeper     commonClients.Gatekeeper
		seagull        commonClients.Seagull
		metrics        highwater.Client
		Config         Config
		LanguageBundle *i18n.Bundle
	}
	Config struct {
		ServerSecret string `json:"serverSecret"` //used for services
		WebURL       string `json:"webUrl"`
		AssetURL     string `json:"assetUrl"`
	}

	group struct {
		Members []string
	}
	// this just makes it easier to bind a handler for the Handle function
	varsHandler func(http.ResponseWriter, *http.Request, map[string]string)
)

const (
	TP_SESSION_TOKEN = "x-tidepool-session-token"

	//returned error messages
	STATUS_ERR_SENDING_EMAIL         = "Error sending email"
	STATUS_ERR_SAVING_CONFIRMATION   = "Error saving the confirmation"
	STATUS_ERR_CREATING_CONFIRMATION = "Error creating a confirmation"
	STATUS_ERR_FINDING_CONFIRMATION  = "Error finding the confirmation"
	STATUS_ERR_FINDING_USER          = "Error finding the user"
	STATUS_ERR_DECODING_CONFIRMATION = "Error decoding the confirmation"
	STATUS_ERR_FINDING_PREVIEW       = "Error finding the invite preview"

	//returned status messages
	STATUS_NOT_FOUND     = "Nothing found"
	STATUS_NO_TOKEN      = "No x-tidepool-session-token was found"
	STATUS_INVALID_TOKEN = "The x-tidepool-session-token was invalid"
	STATUS_UNAUTHORIZED  = "Not authorized for requested operation"
	STATUS_OK            = "OK"
)

func InitApi(
	cfg Config,
	store clients.StoreClient,
	ntf clients.Notifier,
	sl shoreline.Client,
	gatekeeper commonClients.Gatekeeper,
	metrics highwater.Client,
	seagull commonClients.Seagull,
	templates models.Templates,
) *Api {
	locBundle := initI18n()

	return &Api{
		Store:          store,
		Config:         cfg,
		notifier:       ntf,
		sl:             sl,
		gatekeeper:     gatekeeper,
		metrics:        metrics,
		seagull:        seagull,
		templates:      templates,
		LanguageBundle: locBundle,
	}
}

func (a *Api) SetHandlers(prefix string, rtr *mux.Router) {

	rtr.HandleFunc("/status", a.GetStatus).Methods("GET")

	// POST /confirm/send/signup/:userid
	// POST /confirm/send/forgot/:useremail
	// POST /confirm/send/invite/:userid
	send := rtr.PathPrefix("/send").Subrouter()
	send.Handle("/signup/{userid}", varsHandler(a.sendSignUp)).Methods("POST")
	send.Handle("/forgot/{useremail}", varsHandler(a.passwordReset)).Methods("POST")
	send.Handle("/invite/{userid}", varsHandler(a.SendInvite)).Methods("POST")

	// POST /confirm/resend/signup/:useremail
	rtr.Handle("/resend/signup/{useremail}", varsHandler(a.resendSignUp)).Methods("POST")

	// PUT /confirm/accept/signup/:confirmationID
	// PUT /confirm/accept/forgot/
	// PUT /confirm/accept/invite/:userid/:invited_by
	accept := rtr.PathPrefix("/accept").Subrouter()
	accept.Handle("/signup/{confirmationid}", varsHandler(a.acceptSignUp)).Methods("PUT")
	accept.Handle("/forgot", varsHandler(a.acceptPassword)).Methods("PUT")
	accept.Handle("/invite/{userid}/{invitedby}", varsHandler(a.AcceptInvite)).Methods("PUT")

	// GET /confirm/signup/:userid
	// GET /confirm/invite/:userid
	rtr.Handle("/signup/{userid}", varsHandler(a.getSignUp)).Methods("GET")
	rtr.Handle("/invite/{userid}", varsHandler(a.GetSentInvitations)).Methods("GET")

	// GET /confirm/invitations/:userid
	rtr.Handle("/invitations/{userid}", varsHandler(a.GetReceivedInvitations)).Methods("GET")

	// PUT /confirm/dismiss/invite/:userid/:invited_by
	// PUT /confirm/dismiss/signup/:userid
	dismiss := rtr.PathPrefix("/dismiss").Subrouter()
	dismiss.Handle("/invite/{userid}/{invitedby}",
		varsHandler(a.DismissInvite)).Methods("PUT")
	dismiss.Handle("/signup/{userid}",
		varsHandler(a.dismissSignUp)).Methods("PUT")

	// PUT /confirm/:userid/invited/:invited_address
	// PUT /confirm/signup/:userid
	rtr.Handle("/{userid}/invited/{invited_address}", varsHandler(a.CancelInvite)).Methods("PUT")
	rtr.Handle("/signup/{userid}", varsHandler(a.cancelSignUp)).Methods("PUT")
}

func (h varsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(res, req, vars)
}

func (a *Api) GetStatus(res http.ResponseWriter, req *http.Request) {
	if err := a.Store.Ping(); err != nil {
		log.Printf("Error getting status [%v]", err)
		statusErr := &status.StatusError{status.NewStatus(http.StatusInternalServerError, err.Error())}
		a.sendModelAsResWithStatus(res, statusErr, http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(STATUS_OK))
	return
}

//Save this confirmation or
//write an error if it all goes wrong
func (a *Api) addOrUpdateConfirmation(conf *models.Confirmation, res http.ResponseWriter) bool {
	if err := a.Store.UpsertConfirmation(conf); err != nil {
		log.Printf("Error saving the confirmation [%v]", err)
		statusErr := &status.StatusError{status.NewStatus(http.StatusInternalServerError, STATUS_ERR_SAVING_CONFIRMATION)}
		a.sendModelAsResWithStatus(res, statusErr, http.StatusInternalServerError)
		return false
	}
	return true
}

//Find this confirmation
//write error if it fails
func (a *Api) findExistingConfirmation(conf *models.Confirmation, res http.ResponseWriter) (*models.Confirmation, error) {
	if found, err := a.Store.FindConfirmation(conf); err != nil {
		log.Printf("findExistingConfirmation: [%v]", err)
		statusErr := &status.StatusError{status.NewStatus(http.StatusInternalServerError, STATUS_ERR_FINDING_CONFIRMATION)}
		return nil, statusErr
	} else {
		return found, nil
	}
}

//Find this confirmation
//write error if it fails
func (a *Api) addProfile(conf *models.Confirmation) error {
	if conf.CreatorId != "" {
		if err := a.seagull.GetCollection(conf.CreatorId, "profile", a.sl.TokenProvide(), &conf.Creator.Profile); err != nil {
			log.Printf("error getting the creators profile [%v] ", err)
			return err
		}

		conf.Creator.UserId = conf.CreatorId
	}
	return nil
}

//Find these confirmations
//write error if fails or write no-content if it doesn't exist
func (a *Api) checkFoundConfirmations(res http.ResponseWriter, results []*models.Confirmation, err error) []*models.Confirmation {
	if err != nil {
		log.Println("Error finding confirmations ", err)
		statusErr := &status.StatusError{status.NewStatus(http.StatusInternalServerError, STATUS_ERR_FINDING_CONFIRMATION)}
		a.sendModelAsResWithStatus(res, statusErr, http.StatusInternalServerError)
		return nil
	} else if results == nil || len(results) == 0 {
		statusErr := &status.StatusError{status.NewStatus(http.StatusNotFound, STATUS_NOT_FOUND)}
		log.Println("No confirmations were found ", statusErr.Error())
		a.sendModelAsResWithStatus(res, statusErr, http.StatusNotFound)
		return nil
	} else {
		for i := range results {
			if err = a.addProfile(results[i]); err != nil {
				//report and move on
				log.Println("Error getting profile", err.Error())
			}
		}
		return results
	}
}

//Generate a notification from the given confirmation,write the error if it fails
func (a *Api) createAndSendNotification(conf *models.Confirmation, content map[string]interface{}) bool {

	lang := getUserLanguage(conf, a)

	// Get the template name based on the requested communication type
	templateName := conf.TemplateName
	if templateName == models.TemplateNameUndefined {
		switch conf.Type {
		case models.TypePasswordReset:
			templateName = models.TemplateNamePasswordReset
		case models.TypeCareteamInvite:
			templateName = models.TemplateNameCareteamInvite
		case models.TypeSignUp:
			templateName = models.TemplateNameSignup
		case models.TypeNoAccount:
			templateName = models.TemplateNameNoAccount
		default:
			log.Printf("Unknown confirmation type %s", conf.Type)
			return false
		}
	}

	// Content collection is here to replace placeholders in template body/content
	content["WebURL"] = a.Config.WebURL
	content["AssetURL"] = a.Config.AssetURL

	// Retrieve the template from all the preloaded templates
	template, ok := a.templates[templateName]
	if !ok {
		log.Printf("Unknown template type %s", templateName)
		return false
	}

	// Add dynamic content to the template
	fillTemplate(template, a.LanguageBundle, lang, content)

	// Email information (subject and body) are retrieved from the "executed" email template
	// "Execution" adds dynamic content using text/template lib
	subject, body, err := template.Execute(content)
	if err != nil {
		log.Printf("Error executing email template %s", err)
		return false
	}
	// Get localized subject of email
	subject, err = getLocalizedSubject(a.LanguageBundle, subject, lang)

	// Finally send the email
	if status, details := a.notifier.Send([]string{conf.Email}, subject, body); status != http.StatusOK {
		log.Printf("Issue sending email: Status [%d] Message [%s]", status, details)
		return false
	}
	return true
}

//find and validate the token
func (a *Api) token(res http.ResponseWriter, req *http.Request) *shoreline.TokenData {
	if token := req.Header.Get(TP_SESSION_TOKEN); token != "" {
		td := a.sl.CheckToken(token)

		if td == nil {
			statusErr := &status.StatusError{Status: status.NewStatus(http.StatusForbidden, STATUS_INVALID_TOKEN)}
			log.Printf("token %s err[%v] ", STATUS_INVALID_TOKEN, statusErr)
			a.sendModelAsResWithStatus(res, statusErr, http.StatusForbidden)
			return nil
		}
		//all good!
		return td
	}
	statusErr := &status.StatusError{Status: status.NewStatus(http.StatusUnauthorized, STATUS_NO_TOKEN)}
	log.Printf("token %s err[%v] ", STATUS_NO_TOKEN, statusErr)
	a.sendModelAsResWithStatus(res, statusErr, http.StatusUnauthorized)
	return nil
}

//send metric
func (a *Api) logMetric(name string, req *http.Request) {
	token := req.Header.Get(TP_SESSION_TOKEN)
	emptyParams := make(map[string]string)
	a.metrics.PostThisUser(name, token, emptyParams)
	return
}

//send metric
func (a *Api) logMetricAsServer(name string) {
	token := a.sl.TokenProvide()
	emptyParams := make(map[string]string)
	a.metrics.PostServer(name, token, emptyParams)
	return
}

//Find existing user based on the given indentifier
//The indentifier could be either an id or email address
func (a *Api) findExistingUser(indentifier, token string) *shoreline.UserData {
	if usr, err := a.sl.GetUser(indentifier, token); err != nil {
		log.Printf("Error [%s] trying to get existing users details", err.Error())
		return nil
	} else {
		return usr
	}
}

//Makesure we have set the userId on these confirmations
func (a *Api) ensureIdSet(userId string, confirmations []*models.Confirmation) {

	if len(confirmations) < 1 {
		return
	}
	for i := range confirmations {
		//set the userid if not set already
		if confirmations[i].UserId == "" {
			log.Println("UserId wasn't set for invite so setting it")
			confirmations[i].UserId = userId
			a.Store.UpsertConfirmation(confirmations[i])
		}
		return
	}
}

func (a *Api) sendModelAsResWithStatus(res http.ResponseWriter, model interface{}, statusCode int) {
	if jsonDetails, err := json.Marshal(model); err != nil {
		log.Printf("Error [%s] trying to send model [%s]", err.Error(), model)
		http.Error(res, "Error marshaling data for response", http.StatusInternalServerError)
	} else {
		res.Header().Set("content-type", "application/json")
		res.WriteHeader(statusCode)
		res.Write(jsonDetails)
	}
	return
}

func (a *Api) sendError(res http.ResponseWriter, statusCode int, reason string, extras ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		segments := strings.Split(file, "/")
		file = segments[len(segments)-1]
	} else {
		file = "???"
		line = 0
	}

	messages := make([]string, len(extras))
	for index, extra := range extras {
		messages[index] = fmt.Sprintf("%v", extra)
	}

	log.Printf("%s:%d RESPONSE ERROR: [%d %s] %s", file, line, statusCode, reason, strings.Join(messages, "; "))
	a.sendModelAsResWithStatus(res, status.NewStatus(statusCode, reason), statusCode)
}

func (a *Api) sendErrorWithCode(res http.ResponseWriter, statusCode int, errorCode int, reason string, extras ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		segments := strings.Split(file, "/")
		file = segments[len(segments)-1]
	} else {
		file = "???"
		line = 0
	}

	messages := make([]string, len(extras))
	for index, extra := range extras {
		messages[index] = fmt.Sprintf("%v", extra)
	}

	log.Printf("%s:%d RESPONSE ERROR: [%d %s] %s", file, line, statusCode, reason, strings.Join(messages, "; "))
	a.sendModelAsResWithStatus(res, status.NewStatusWithError(statusCode, errorCode, reason), statusCode)
}

func (a *Api) tokenUserHasRequestedPermissions(tokenData *shoreline.TokenData, groupId string, requestedPermissions commonClients.Permissions) (commonClients.Permissions, error) {
	if tokenData.IsServer {
		return requestedPermissions, nil
	} else if tokenData.UserID == groupId {
		return requestedPermissions, nil
	} else if actualPermissions, err := a.gatekeeper.UserInGroup(tokenData.UserID, groupId); err != nil {
		return commonClients.Permissions{}, err
	} else {
		finalPermissions := make(commonClients.Permissions, 0)
		for permission, _ := range requestedPermissions {
			if reflect.DeepEqual(requestedPermissions[permission], actualPermissions[permission]) {
				finalPermissions[permission] = requestedPermissions[permission]
			}
		}
		return finalPermissions, nil
	}
}

// getAllLocalizationFiles returns all the filenames within the templates/locales folder
// Add yaml file to this folder to get a language added
// At least en.yaml should be present
func getAllLocalizationFiles() ([]string, error) {
	var files []string
	// Walk the folder and add files one by one
	err := filepath.Walk("templates/locales", func(path string, info os.FileInfo, err error) error {
		// All files not directory
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	// Return all files
	return files, err
}

// initI18n initializes the internationalization objects needed by the application
// Ensure at least en.yaml is present in the "templates/locales" folder
func initI18n() *i18n.Bundle {
	// Get all the language files that exist
	langFiles, err := getAllLocalizationFiles()

	if err != nil {
		fmt.Errorf("Error getting translation files, %v", err)
		panic(err)
	}

	// Create a Bundle to use for the lifetime of your application
	locBundle, err := createLocalizerBundle(langFiles)

	if err != nil {
		fmt.Errorf("Error initialising localization, %v", err)
		panic(err)
	}

	return locBundle
}

// createLocalizerBundle reads language files and registers them in i18n bundle
func createLocalizerBundle(langFiles []string) (*i18n.Bundle, error) {
	// Bundle stores a set of messages
	bundle := &i18n.Bundle{DefaultLanguage: language.English}

	// Enable bundle to understand yaml
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	var translations []byte
	var err error
	for _, file := range langFiles {

		// Read our language yaml file
		translations, err = ioutil.ReadFile(file)
		if err != nil {
			fmt.Errorf("Unable to read translation file %s", file)
			return nil, err
		}

		// It parses the bytes in buffer to add translations to the bundle
		bundle.MustParseMessageFileBytes(translations, file)
	}

	return bundle, nil
}

// getLocalizedContentPart returns translated content part based on key and locale
func getLocalizedContentPart(bundle *i18n.Bundle, key string, locale string, escape map[string]interface{}) (string, error) {
	localizer := i18n.NewLocalizer(bundle, locale)
	msg, err := localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: escape,
		},
	)
	if msg == "" {
		msg = "<< Cannot find translation for item " + key + " >>"
	}
	return msg, err
}

// getLocalizedSubject returns translated subject based on key and locale
func getLocalizedSubject(bundle *i18n.Bundle, key string, locale string) (string, error) {
	return getLocalizedContentPart(bundle, key, locale, nil)
}

// fillTemplate fills the template content parts based on language bundle and locale
// A template content/body is made of HTML tags and content that can be localized
// Each template references its parts that can be filled in a collection called ContentParts
func fillTemplate(template models.Template, bundle *i18n.Bundle, locale string, content map[string]interface{}) {
	// Get content parts from the template
	for _, v := range template.ContentParts() {
		// Each part is translated in the requested locale and added to the Content collection
		contentItem, _ := getLocalizedContentPart(bundle, v, locale, fillEscapedParts(template, content))
		content[v] = contentItem
	}
}

// fillEscapedParts dynamically fills the escape parts with content
func fillEscapedParts(template models.Template, content map[string]interface{}) map[string]interface{} {

	// Escaped parts are replaced with content value
	var escape = make(map[string]interface{})
	if template.EscapeParts() != nil {
		for _, v := range template.EscapeParts() {
			escape[v] = content[v]
		}
	}

	return escape
}

// getUserLanguage returns the language of the user
func getUserLanguage(conf *models.Confirmation, a *Api) string {
	// Get user profile and language for message
	type (
		UserProfile struct {
			Language string `json:"language"`
		}
	)

	var profile = &UserProfile{}
	a.seagull.GetCollection(conf.UserId, "profile", a.sl.TokenProvide(), profile)

	if profile.Language == "" {
		profile.Language = "en"
	}

	return profile.Language
}

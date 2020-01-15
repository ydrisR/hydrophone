package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/tidepool-org/go-common/clients/status"
	"github.com/tidepool-org/hydrophone/localize"
	"github.com/tidepool-org/hydrophone/models"
	"github.com/tidepool-org/hydrophone/templates"
)

type (
	Api struct {
		Config        Config
		templates     models.Templates
		localeManager LocaleManager
	}
	// this just makes it easier to bind a handler for the Handle function
	varsHandler func(http.ResponseWriter, *http.Request, map[string]string)
)

// Init the preview api with configuration
func InitApi(
	cfg Config,
	templates models.Templates,
	localeManager LocaleManager,
) *Api {
	return &Api{
		Config:        cfg,
		templates:     templates,
		localeManager: localeManager,
	}
}

func (a *Api) SetHandlers(prefix string, rtr *mux.Router) {
	rtr.Handle("/preview/{template}", varsHandler(a.preview)).Methods("GET")
	rtr.Handle("/refreshlocal", varsHandler(a.refreshLocal)).Methods("POST")
	rtr.HandleFunc("/", a.serveStatic).Methods("GET")
}

func (h varsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(res, req, vars)
}

// Return the index page
func (a *Api) serveStatic(res http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadFile("index.html")
	if err != nil {
		log.Printf("templates - failure to read index.html")
	}
	res.Header().Set("content-type", "text/html")
	res.WriteHeader(200)
	res.Write(data)
	return
}

// Refresh locales from the localization system (i.e. Loco)
func (a *Api) refreshLocal(res http.ResponseWriter, req *http.Request, vars map[string]string) {
	log.Println("refresh locales files from remote system!")
	success := a.localeManager.DownloadLocales("../locales/")
	if !success {
		log.Printf("error while retriving locales from localizer service!")
		s := status.NewApiStatus(http.StatusInternalServerError, "error")
		a.sendModelAsResWithStatus(res, s, http.StatusInternalServerError)
	}
	localizer, err := localize.NewI18nLocalizer("../locales/")
	if err != nil {
		log.Printf("error while reloading locales files!")
		s := status.NewApiStatus(http.StatusInternalServerError, "error")
		a.sendModelAsResWithStatus(res, s, http.StatusInternalServerError)
	}
	emailTemplates, err := templates.New("..", localizer)
	if err != nil {
		log.Printf("error while reloading templates!")
		s := status.NewApiStatus(http.StatusInternalServerError, "error")
		a.sendModelAsResWithStatus(res, s, http.StatusInternalServerError)
	}
	a.templates = emailTemplates
}

// Compile a template with test content and return the html result
func (a *Api) preview(res http.ResponseWriter, req *http.Request, vars map[string]string) {
	//Determine the email template:
	var templateName models.TemplateName
	lang := "en"
	switch vars["template"] {
	case "careteam_invitation":
		templateName = models.TemplateNameCareteamInvite
	case "no_account":
		templateName = models.TemplateNameNoAccount
	case "password_reset":
		templateName = models.TemplateNamePasswordReset
	case "patient_password_reset":
		templateName = models.TemplateNamePatientPasswordReset
	case "patient_information":
		templateName = models.TemplateNamePatientInformation
	case "signup_confirmation":
		templateName = models.TemplateNameSignup
	case "signup_clinic_confirmation":
		templateName = models.TemplateNameSignupClinic
	case "signup_custodial_confirmation":
		templateName = models.TemplateNameSignupCustodial
	case "signup_custodial_clinic_confirmation":
		templateName = models.TemplateNameSignupCustodialClinic
	default:
		log.Printf("Unknown template %s", vars["template"])
		s := status.NewApiStatus(400, "Incorrect template name")
		a.sendModelAsResWithStatus(res, s, http.StatusInternalServerError)
		return
	}

	langs, ok := req.URL.Query()["lang"]
	if ok && len(langs[0]) == 2 {
		lang = langs[0]
	}
	email, err := a.generateEmail(templateName, lang)
	if err != nil {
		log.Println("Error generating email preview", err)
		s := status.NewApiStatus(http.StatusInternalServerError, err.Error())
		a.sendModelAsResWithStatus(res, s, http.StatusInternalServerError)
	} else {
		res.Header().Set("content-type", "text/html")
		res.WriteHeader(200)
		res.Write([]byte(email))
	}

	return
}

//Generate a notification from the given confirmation,write the error if it fails
func (a *Api) generateEmail(templateName models.TemplateName, lang string) (string, error) {

	log.Printf("trying preview with template '%s' with language '%s'", templateName, lang)

	content := map[string]interface{}{
		"Key":          "123456789",
		"Email":        "john@diabeloop.com",
		"FullName":     "John Doe",
		"CareteamName": "John Doe",
		"WebPath":      "login",
	}
	// Content collection is here to replace placeholders in template body/content
	content["CreatorName"] = "John Doe"
	content["WebURL"] = a.Config.WebURL
	content["SupportURL"] = a.Config.SupportURL
	content["AssetURL"] = a.Config.AssetURL
	content["PatientPasswordResetURL"] = a.Config.PatientPasswordResetURL

	// Retrieve the template from all the preloaded templates

	template, ok := a.templates[templateName]
	if !ok {
		return "", fmt.Errorf("Unknown template type %s", templateName)
	}

	// Get localized subject of email
	subject, body, err := template.Execute(content, lang)
	if err != nil {
		return "", fmt.Errorf("Error executing email template '%s'", err)
	}
	result := fmt.Sprintf("<div align=\"center\" id=\"subject\">Subject: %s \n</div><div id=\"body\">%s</div>", subject, body)
	return result, nil
}

func (a *Api) sendModelAsResWithStatus(res http.ResponseWriter, model interface{}, statusCode int) {
	if jsonDetails, err := json.Marshal(model); err != nil {
		log.Printf("Error [%s] trying to preview model [%s]", err.Error(), model)
		http.Error(res, "Error marshaling data for response", http.StatusInternalServerError)
	} else {
		res.Header().Set("content-type", "application/json")
		res.WriteHeader(statusCode)
		res.Write(jsonDetails)
	}
	return
}

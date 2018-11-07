package templates

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tidepool-org/hydrophone/models"
)

type TemplateMeta struct {
	Name                    string   `json:"name"`
	Description             string   `json:"description"`
	HTMLPath                string   `json:"htmlPath"`
	ContentChunks           []string `json:"contentChunks"`
	Subject                 string   `json:"subject"`
	EscapeTranslationChunks []string `json:"escapeTranslationChunks"`
}

// getTemplateMeta returns the template metadata
// Metadata are information that relate to a template (e.g. name, htmlPath...)
// Inputs:
// metaFileName = name of the file with no path and no json extension, assuming the file is located in templates/meta
func getTemplateMeta(metaFileName string) TemplateMeta {
	// Open the jsonFile
	jsonFile, err := os.Open("templates/meta/" + metaFileName + ".json")
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read the opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var meta TemplateMeta

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &meta)

	return meta
}

// getBody returns the email body corresponding to the template requested
func getBody(t string) string {
	dat, _ := ioutil.ReadFile(t)
	return string(dat)
}

func New() (models.Templates, error) {
	templates := models.Templates{}

	if template, err := NewCareteamInviteTemplate(); err != nil {
		return nil, fmt.Errorf("templates: failure to create careteam invite template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewNoAccountTemplate(); err != nil {
		return nil, fmt.Errorf("templates: failure to create no account template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewPasswordResetTemplate(); err != nil {
		return nil, fmt.Errorf("templates: failure to create password reset template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewSignupTemplate(); err != nil {
		return nil, fmt.Errorf("templates: failure to create signup template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewSignupClinicTemplate(); err != nil {
		return nil, fmt.Errorf("templates: failure to create signup clinic template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewSignupCustodialTemplate(); err != nil {
		return nil, fmt.Errorf("templates: failure to create signup custodial template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewSignupCustodialClinicTemplate(); err != nil {
		return nil, fmt.Errorf("templates: failure to create signup custodial clinic template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	return templates, nil
}

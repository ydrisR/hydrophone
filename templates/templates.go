package templates

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/tidepool-org/hydrophone/models"
)

type TemplateMeta struct {
	Name                    string   `json:"name"`
	Description             string   `json:"description"`
	TemplateFilename        string   `json:"templateFilename"`
	ContentChunks           []string `json:"contentChunks"`
	Subject                 string   `json:"subject"`
	EscapeTranslationChunks []string `json:"escapeTranslationChunks"`
}

// getTemplateMeta returns the template metadata
// Metadata are information that relate to a template (e.g. name, templateFilename...)
// Inputs:
// metaFileName = name of the file with no path and no json extension, assuming the file is located in path specified in TIDEPOOL_HYDROPHONE_SERVICE environment variable
func getTemplateMeta(metaFileName string) TemplateMeta {
	log.Printf("getting template meta from %s", metaFileName)

	// Open the jsonFile
	jsonFile, err := os.Open(metaFileName)
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

// getBody returns the email body from the file which name is in input parameter
func getBody(fileName string) string {

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("templates - failure to get template body: %s", err)
	}
	log.Printf("getting template body from %s", fileName)
	return string(data)
}

func New(templatesPath string) (models.Templates, error) {
	templates := models.Templates{}

	if template, err := NewCareteamInviteTemplate(templatesPath); err != nil {
		return nil, fmt.Errorf("templates: failure to create careteam invite template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewNoAccountTemplate(templatesPath); err != nil {
		return nil, fmt.Errorf("templates: failure to create no account template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewPasswordResetTemplate(templatesPath); err != nil {
		return nil, fmt.Errorf("templates: failure to create password reset template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewSignupTemplate(templatesPath); err != nil {
		return nil, fmt.Errorf("templates: failure to create signup template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewSignupClinicTemplate(templatesPath); err != nil {
		return nil, fmt.Errorf("templates: failure to create signup clinic template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewSignupCustodialTemplate(templatesPath); err != nil {
		return nil, fmt.Errorf("templates: failure to create signup custodial template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	if template, err := NewSignupCustodialClinicTemplate(templatesPath); err != nil {
		return nil, fmt.Errorf("templates: failure to create signup custodial clinic template: %s", err)
	} else {
		templates[template.Name()] = template
	}

	return templates, nil
}

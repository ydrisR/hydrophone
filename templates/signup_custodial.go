package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupCustodialTemplate(templatesPath string) (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta(templatesPath + "/meta/signup_custodial_confirmation.json")
	var templateFileName = templatesPath + "/html/" + templateMeta.TemplateFilename

	return models.NewPrecompiledTemplate(models.TemplateNameSignupCustodial, templateMeta.Subject, getBody(templateFileName), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

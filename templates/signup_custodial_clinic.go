package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupCustodialClinicTemplate(templatesPath string) (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta(templatesPath + "/meta/signup_custodial_clinic_confirmation.json")
	var templateFileName = templatesPath + "/html/" + templateMeta.TemplateFilename

	return models.NewPrecompiledTemplate(models.TemplateNameSignupCustodialClinic, templateMeta.Subject, getBody(templateFileName), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

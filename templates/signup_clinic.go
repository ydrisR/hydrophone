package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupClinicTemplate(templatesPath string) (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta(templatesPath + "/meta/signup_clinic_confirmation.json")
	var templateFileName = templatesPath + "/html/" + templateMeta.TemplateFilename

	return models.NewPrecompiledTemplate(models.TemplateNameSignupClinic, templateMeta.Subject, getBody(templateFileName), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

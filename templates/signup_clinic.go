package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupClinicTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = GetTemplateMeta("signup_clinic")

	return models.NewPrecompiledTemplate(models.TemplateNameSignupClinic, templateMeta.Subject, GetBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

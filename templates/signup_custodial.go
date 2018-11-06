package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupCustodialTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta("signup_custodial")

	return models.NewPrecompiledTemplate(models.TemplateNameSignupCustodial, templateMeta.Subject, getBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

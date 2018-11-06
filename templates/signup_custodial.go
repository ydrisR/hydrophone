package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupCustodialTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = GetTemplateMeta("signup_custodial")

	return models.NewPrecompiledTemplate(models.TemplateNameSignupCustodial, templateMeta.Subject, GetBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

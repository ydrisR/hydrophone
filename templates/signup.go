package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = GetTemplateMeta("signup")

	return models.NewPrecompiledTemplate(models.TemplateNameSignup, templateMeta.Subject, GetBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

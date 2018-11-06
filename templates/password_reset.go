package templates

import "github.com/tidepool-org/hydrophone/models"

// NewPasswordResetTemplate sends template
func NewPasswordResetTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta("password_reset")

	return models.NewPrecompiledTemplate(models.TemplateNamePasswordReset, templateMeta.Subject, getBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

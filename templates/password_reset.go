package templates

import "github.com/tidepool-org/hydrophone/models"

// NewPasswordResetTemplate sends template
func NewPasswordResetTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = GetTemplateMeta("password_reset")

	return models.NewPrecompiledTemplate(models.TemplateNamePasswordReset, templateMeta.Subject, GetBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

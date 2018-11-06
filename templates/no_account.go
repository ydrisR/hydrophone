package templates

import (
	"github.com/tidepool-org/hydrophone/models"
)

func NewNoAccountTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta("no_account")

	return models.NewPrecompiledTemplate(models.TemplateNameNoAccount, templateMeta.Subject, getBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

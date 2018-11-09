package templates

import (
	"github.com/tidepool-org/hydrophone/models"
)

func NewNoAccountTemplate(templatesPath string) (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta(templatesPath + "/meta/no_account.json")
	var templateFileName = templatesPath + "/html/" + templateMeta.TemplateFilename

	return models.NewPrecompiledTemplate(models.TemplateNameNoAccount, templateMeta.Subject, getBody(templateFileName), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

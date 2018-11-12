package templates

import (
	"github.com/tidepool-org/hydrophone/models"
)

func NewTestTemplate(templatesPath string) (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta(templatesPath + "/meta/test_template.json")
	var templateFileName = templatesPath + "/html/" + templateMeta.TemplateFilename

	return models.NewPrecompiledTemplate(models.TemplateNameTestTemplate, templateMeta.Subject, getBody(templateFileName), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

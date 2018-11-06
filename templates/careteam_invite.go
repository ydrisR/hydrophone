package templates

import (
	"github.com/tidepool-org/hydrophone/models"
)

func NewCareteamInviteTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta("careteam_invite")

	return models.NewPrecompiledTemplate(models.TemplateNameCareteamInvite, templateMeta.Subject, getBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

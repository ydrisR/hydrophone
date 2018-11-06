package templates

import (
	"github.com/tidepool-org/hydrophone/models"
)

func NewCareteamInviteTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = GetTemplateMeta("careteam_invite")

	return models.NewPrecompiledTemplate(models.TemplateNameCareteamInvite, templateMeta.Subject, GetBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

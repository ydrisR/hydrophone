package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupCustodialClinicTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = getTemplateMeta("signup_custodial_clinic")

	return models.NewPrecompiledTemplate(models.TemplateNameSignupCustodialClinic, templateMeta.Subject, getBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

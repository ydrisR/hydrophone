package templates

import "github.com/tidepool-org/hydrophone/models"

func NewSignupCustodialClinicTemplate() (models.Template, error) {

	// Get template Metadata
	var templateMeta = GetTemplateMeta("signup_custodial_clinic")

	return models.NewPrecompiledTemplate(models.TemplateNameSignupCustodialClinic, templateMeta.Subject, GetBody(templateMeta.HTMLPath), templateMeta.ContentChunks, templateMeta.EscapeTranslationChunks)
}

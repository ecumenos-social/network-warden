package emailer

import (
	"fmt"
	"html/template"
	"path/filepath"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
)

type TemplateName string

const (
	TemplateNameConfirmHolderRegistration TemplateName = "confirm-holder-registration"
)

var unknownTemplateName = func(tn TemplateName) error {
	return errorwrapper.New(fmt.Sprintf("unknown template name, name = %v", tn))
}

func (tn TemplateName) Validate() error {
	for _, n := range []TemplateName{TemplateNameConfirmHolderRegistration} {
		if n == tn {
			return nil
		}
	}
	return unknownTemplateName(tn)
}

func takeTemplate(name TemplateName) (*template.Template, error) {
	if err := name.Validate(); err != nil {
		return nil, errorwrapper.NewWithError(err)
	}
	path := fmt.Sprintf("services/emailer/templates/%s.html", name)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	t, err := template.ParseFiles(absPath)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func takeSubject(name TemplateName) (string, error) {
	switch name {
	case TemplateNameConfirmHolderRegistration:
		return "Confirmation of Registration", nil
	}

	return "", unknownTemplateName(name)
}

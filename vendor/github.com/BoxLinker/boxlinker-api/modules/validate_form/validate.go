package validate_form

import "github.com/Sirupsen/logrus"

type Form interface{
	Validate() error
}

type ValidateForm struct {

}

func (form *ValidateForm) Validate(names ...string) error {
	logrus.Debugf("form:> %+v", form)
	return nil
}


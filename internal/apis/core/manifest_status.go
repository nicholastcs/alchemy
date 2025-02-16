package core

import "errors"

const ResourceReady string = "ResourceReady"

type Status struct {
	Conditions []Condition `yaml:"conditions" mapstructure:"conditions" json:"conditions"`
	Errors     []Error     `yaml:"errors,omitempty" mapstructure:"errors,omitempty" json:"errors,omitempty"`
}

type Condition struct {
	Type   string `yaml:"type" mapstructure:"type" json:"type"`
	Status bool   `yaml:"status" mapstructure:"status" json:"status"`
}

type Error struct {
	Message string `yaml:"message" mapstructure:"message" json:"error"`
}

func NewStatus() (*Status, error) {
	return &Status{
		Conditions: []Condition{},
		Errors:     []Error{},
	}, nil
}

func (s *Status) SetCondition(typeName string, status bool) {
	if typeName == "" {
		panic("type name cannot be empty")
	}

	for i, c := range s.Conditions {
		if c.Type == typeName {
			s.Conditions[i].Status = status

			return
		}
	}

	s.Conditions = append(s.Conditions, Condition{Type: typeName, Status: status})
}

func (s *Status) GetCondition(typeName string) bool {
	for _, c := range s.Conditions {
		if c.Type == typeName {
			return c.Status
		}
	}

	s.Conditions = append(s.Conditions, Condition{Type: typeName, Status: false})
	return false
}

func (s *Status) SetError(err error) {
	if err == nil {
		return
	}

	for _, e := range s.Errors {
		if e.Message == err.Error() {
			return
		}
	}
	s.Errors = append(s.Errors, Error{Message: err.Error()})
}

func (s *Status) HasErr() bool {
	return len(s.Errors) > 0
}

func (s *Status) ToNativeErr() error {
	if !s.HasErr() {
		return nil
	}
	var errs error
	for _, e := range s.Errors {
		errs = errors.Join(errs, errors.New(e.Message))
	}
	return errs
}

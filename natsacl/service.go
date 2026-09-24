package natsacl

import "fmt"

// ServiceBuilder is used to generate micro services related permissions
type ServiceBuilder struct {
	parent   *Builder
	name     string
	subjects []string
}

// Service returns an service builder
func (b *Builder) Service(name string, subjects ...string) *ServiceBuilder {
	return &ServiceBuilder{
		parent:   b,
		name:     name,
		subjects: subjects,
	}
}

// Build returns the parent builder
func (s *ServiceBuilder) Build() *Builder {
	return s.parent
}

// Request allows to send a request to a service
func (s *ServiceBuilder) Request() *ServiceBuilder {
	for _, subject := range s.subjects {
		s.parent.AllowPub(subject)
	}
	s.parent.inbox()
	return s
}

// Respond allows to respond to requests
func (s *ServiceBuilder) Respond() *ServiceBuilder {
	for _, subject := range s.subjects {
		s.parent.AllowSub(subject)
	}
	s.parent.pubInbox()
	return s
}

// Monitor allows requests to the services monitoring API
func (s *ServiceBuilder) Monitor() *ServiceBuilder {
	if s.name == "*" {
		s.parent.AllowPub("$SRV.PING")
		s.parent.AllowPub("$SRV.INFO")
		s.parent.AllowPub("$SRV.STATS")
	}
	for _, api := range []string{"PING", "INFO", "STATS"} {
		s.parent.AllowPub(fmt.Sprintf("$SRV.%s.%s", api, s.name))
		s.parent.AllowPub(fmt.Sprintf("$SRV.%s.%s.*", api, s.name))
	}
	s.parent.inbox()
	return s
}

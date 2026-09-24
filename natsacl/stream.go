package natsacl

import "fmt"

// StreamBuilder is used to build streams related permissions
type StreamBuilder struct {
	parent *Builder
	name   string
}

// Streams returns a stream builder
func (b *Builder) Stream(name string) *StreamBuilder {
	return &StreamBuilder{
		parent: b,
		name:   name,
	}
}

// Build returns the parent builder
func (s *StreamBuilder) Build() *Builder {
	return s.parent
}

// All gives all permissions except List
func (s *StreamBuilder) All() *StreamBuilder {
	return s.List().Consume().Create().Update().Delete().Info().GetMessage().DeleteMessage().Purge()
}

// List gives the permissions to list the streams
func (s *StreamBuilder) List() *StreamBuilder {
	s.parent.list()
	return s
}

// Consume gives the permissions to consume messages from a stream
func (s *StreamBuilder) Consume() *StreamBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", s.name),
		fmt.Sprintf("$JS.API.CONSUMER.CREATE.%s.>", s.name),
		fmt.Sprintf("$JS.API.CONSUMER.DELETE.%s.>", s.name),
		fmt.Sprintf("$JS.API.CONSUMER.MSG.NEXT.%s.>", s.name),
		fmt.Sprintf("$JS.ACK.%s.>", s.name),
	)
	s.parent.inbox()
	return s
}

// Create gives the permissions to create a stream
func (s *StreamBuilder) Create() *StreamBuilder {
	s.parent.allowPub(
		"$JS.API.INFO",
		fmt.Sprintf("$JS.API.STREAM.CREATE.%s", s.name),
	)
	s.parent.inbox()
	return s
}

// Update gives the permissions to update a stream
func (s *StreamBuilder) Update() *StreamBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.UPDATE.%s", s.name),
	)
	return s
}

// Update gives the permissions to delete a stream
func (s *StreamBuilder) Delete() *StreamBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.DELETE.%s", s.name),
	)
	return s
}

// Info gives the permissions to view information on a stream
func (s *StreamBuilder) Info() *StreamBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", s.name),
	)
	s.parent.inbox()
	return s
}

// Get gives the permissions to get messages from a stream
func (s *StreamBuilder) GetMessage() *StreamBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.MSG.GET.%s", s.name),
	)
	return s
}

// Delete gives the permissions to delete messages from a stream
func (s *StreamBuilder) DeleteMessage() *StreamBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.MSG.DELETE.%s", s.name),
	)
	return s
}

// Purge gives the permissions to purge messages from a stream
func (s *StreamBuilder) Purge() *StreamBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.PURGE.%s", s.name),
	)
	return s
}

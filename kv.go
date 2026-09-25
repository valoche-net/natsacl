package natsacl

import "fmt"

// KVBuilder is used to build KV stores related permissions
type KVBuilder struct {
	parent *Builder
	name   string
}

// KV returns a KV store builder
func (b *Builder) KV(name string) *KVBuilder {
	k := KVBuilder{
		parent: b,
		name:   name,
	}
	k.parent.inbox()
	return &k
}

// stream returns the underlying stream name
func (s *KVBuilder) stream() string {
	if s.name == "*" {
		return "*"
	}
	return fmt.Sprintf("KV_%s", s.name)
}

// Build returns the parent builder
func (s *KVBuilder) Build() *Builder {
	return s.parent
}

// All gives all the permissions except list
func (s *KVBuilder) All() *KVBuilder {
	return s.View().Create().Delete().Read().Write()
}

// List gives the permissions to list the store
func (s *KVBuilder) List() *KVBuilder {
	s.parent.names()
	return s
}

// View gives the permissions to view the keys from a store
//
// ⚠️ It also gives the permissions to watch, hence seeing the data
func (s *KVBuilder) View() *KVBuilder {
	s.parent.AllowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", s.stream()),
		fmt.Sprintf("$JS.API.CONSUMER.CREATE.%s.*.$KV.%s.>", s.stream(), s.name),
		fmt.Sprintf("$JS.API.CONSUMER.DELETE.%s.*", s.stream()),
	)
	return s
}

// Create gives the permissions to create a store
func (s *KVBuilder) Create() *KVBuilder {
	s.parent.info()
	s.parent.AllowPub(
		fmt.Sprintf("$JS.API.STREAM.CREATE.%s", s.stream()),
	)
	return s
}

// Delete gives the permissions to delete a store
func (s *KVBuilder) Delete() *KVBuilder {
	s.parent.AllowPub(
		fmt.Sprintf("$JS.API.STREAM.DELETE.%s", s.stream()),
	)
	return s
}

// Read gives the permissions to get data from a store, optionally restricted to certain keys
func (s *KVBuilder) Read(keys ...string) *KVBuilder {
	s.parent.AllowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", s.stream()),
	)
	if len(keys) == 0 {
		s.View()
		s.parent.AllowPub(
			fmt.Sprintf("$JS.API.DIRECT.GET.%s.$KV.%s.*", s.stream(), s.name),
		)
		return s
	}
	for _, key := range keys {
		s.parent.AllowPub(
			fmt.Sprintf("$JS.API.DIRECT.GET.%s.$KV.%s.%s", s.stream(), s.name, key),
		)
	}
	return s
}

// Write gives the permissions to create or update data in a store, optionally restricted to certain keys
func (s *KVBuilder) Write(keys ...string) *KVBuilder {
	s.parent.AllowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", s.stream()),
	)
	if len(keys) == 0 {
		s.parent.AllowPub(
			fmt.Sprintf("$KV.%s.*", s.name),
		)
		return s
	}
	for _, key := range keys {
		s.parent.AllowPub(
			fmt.Sprintf("$KV.%s.%s", s.name, key),
		)
	}
	return s
}

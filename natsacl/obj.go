package natsacl

import (
	"encoding/base64"
	"fmt"
)

// ObjBuilder is used to build stores (KV, obj) related permissions
type ObjBuilder struct {
	parent *Builder
	name   string
}

// Obj returns an object store builder
func (b *Builder) Obj(name string) *ObjBuilder {
	o := ObjBuilder{
		parent: b,
		name:   name,
	}
	o.parent.inbox()
	return &o
}

// stream returns the underlying stream name
func (s *ObjBuilder) stream() string {
	if s.name == "*" {
		return "*"
	}
	return fmt.Sprintf("OBJ_%s", s.name)
}

// Build returns the parent builder
func (s *ObjBuilder) Build() *Builder {
	return s.parent
}

// All gives all the permissions except list
func (s *ObjBuilder) All() *ObjBuilder {
	return s.View().Create().Delete().Read().Write().Seal()
}

// List gives the permissions to list the store
func (s *ObjBuilder) List() *ObjBuilder {
	s.parent.names()
	return s
}

// View gives the permissions to view the keys from a store
//
// ⚠️ It also gives the permissions to watch, hence seeing the data
func (s *ObjBuilder) View() *ObjBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.DIRECT.GET.%s.$O.%s.M.>", s.stream(), s.name),
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", s.stream()),
		fmt.Sprintf("$JS.API.CONSUMER.CREATE.%s.*.$O.%s.M.>", s.stream(), s.name),
		fmt.Sprintf("$JS.API.CONSUMER.DELETE.%s.*", s.stream()),
	)
	return s
}

// Create gives the permissions to create/update a store
func (s *ObjBuilder) Create() *ObjBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.CREATE.%s", s.stream()),
		fmt.Sprintf("$JS.API.STREAM.UPDATE.%s", s.stream()),
	)
	return s
}

// Delete gives the permissions to delete a store
func (s *ObjBuilder) Delete() *ObjBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.DELETE.%s", s.stream()),
	)
	return s
}

// Read gives the permissions to get data from a store, optionally restricted to certain keys
//
// ⚠️ if read is restricted to certain keys, listing the keys is denied
func (s *ObjBuilder) Read(keys ...string) *ObjBuilder {
	if len(keys) == 0 {
		s.parent.allowPub(
			fmt.Sprintf("$JS.API.CONSUMER.CREATE.%s.*.$O.%s.M.*", s.stream(), s.name),
			fmt.Sprintf("$JS.API.DIRECT.GET.%s.$O.%s.M.>", s.stream(), s.name),
		)
	} else {
		for _, key := range keys {
			b64key := base64.StdEncoding.EncodeToString([]byte(key))
			s.parent.allowPub(
				fmt.Sprintf("$JS.API.CONSUMER.CREATE.%s.*.$O.%s.M.%s", s.stream(), s.name, b64key),
				fmt.Sprintf("$JS.API.DIRECT.GET.%s.$O.%s.M.%s", s.stream(), s.name, b64key),
			)
		}

	}
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", s.stream()),
		fmt.Sprintf("$JS.API.CONSUMER.CREATE.%s.*.$O.%s.C.*", s.stream(), s.name),
		fmt.Sprintf("$JS.API.CONSUMER.DELETE.%s.*", s.stream()),
	)
	return s
}

// Write gives the permissions to create or update data in a store, optionally restricted to certain keys
func (s *ObjBuilder) Write(keys ...string) *ObjBuilder {
	if len(keys) == 0 {
		s.parent.allowPub(
			fmt.Sprintf("$JS.API.DIRECT.GET.%s.$O.%s.M.*", s.stream(), s.name),
			fmt.Sprintf("$O.%s.M.*", s.name),
		)
	} else {
		for _, key := range keys {
			b64key := base64.StdEncoding.EncodeToString([]byte(key))
			s.parent.allowPub(
				fmt.Sprintf("$JS.API.DIRECT.GET.%s.$O.%s.M.%s", s.stream(), s.name, b64key),
				fmt.Sprintf("$O.%s.M.%s", s.name, b64key),
			)
		}
	}
	s.parent.allowPub(
		fmt.Sprintf("$O.%s.C.*", s.name),
		fmt.Sprintf("$JS.API.STREAM.PURGE.%s", s.stream()),
	)
	return s
}

// Seal gives the permissions to seal a store
func (s *ObjBuilder) Seal() *ObjBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.%s", s.stream()),
		fmt.Sprintf("$JS.API.STREAM.UPDATE.%s", s.stream()),
	)
	return s
}

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
	s.parent.list()
	return s
}

// View gives the permissions to view the keys from a store
//
// ⚠️ It also gives the permissions to watch, hence seeing the data
func (s *ObjBuilder) View() *ObjBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.$O.%s.M.>", s.name, s.name),
		fmt.Sprintf("$JS.API.STREAM.INFO.OBJ_%s", s.name),
		fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.M.>", s.name, s.name),
		fmt.Sprintf("$JS.API.CONSUMER.DELETE.OBJ_%s.*", s.name),
	)
	return s
}

// Create gives the permissions to create a store
func (s *ObjBuilder) Create() *ObjBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.CREATE.OBJ_%s", s.name),
	)
	return s
}

// Delete gives the permissions to delete a store
func (s *ObjBuilder) Delete() *ObjBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.DELETE.OBJ_%s", s.name),
	)
	return s
}

// Read gives the permissions to get data from a store, optionally restricted to certain keys
//
// ⚠️ if read is restricted to certain keys, listing the keys is denied
func (s *ObjBuilder) Read(keys ...string) *ObjBuilder {
	if len(keys) == 0 {
		s.parent.allowPub(
			fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.M.*", s.name, s.name),
			fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.$O.%s.M.>", s.name, s.name),
		)
	} else {
		for _, key := range keys {
			b64key := base64.StdEncoding.EncodeToString([]byte(key))
			s.parent.allowPub(
				fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.M.%s", s.name, s.name, b64key),
				fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.$O.%s.M.%s", s.name, s.name, b64key),
			)
		}

	}
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.OBJ_%s", s.name),
		fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.C.*", s.name, s.name),
		fmt.Sprintf("$JS.API.CONSUMER.DELETE.OBJ_%s.*", s.name),
	)
	return s
}

// Write gives the permissions to create or update data in a store, optionally restricted to certain keys
func (s *ObjBuilder) Write(keys ...string) *ObjBuilder {
	if len(keys) == 0 {
		s.parent.allowPub(
			fmt.Sprintf("$O.%s.M.*", s.name),
		)
	} else {
		for _, key := range keys {
			b64key := base64.StdEncoding.EncodeToString([]byte(key))
			s.parent.allowPub(
				fmt.Sprintf("$O.%s.M.%s", s.name, b64key),
			)
		}
	}
	s.parent.allowPub(
		fmt.Sprintf("$O.%s.C.*", s.name),
		fmt.Sprintf("$JS.API.STREAM.PURGE.OBJ_%s", s.name),
	)
	return s
}

// Seal gives the permissions to seal a store
func (s *ObjBuilder) Seal() *ObjBuilder {
	s.parent.allowPub(
		fmt.Sprintf("$JS.API.STREAM.INFO.OBJ_%s", s.name),
		fmt.Sprintf("$JS.API.STREAM.UPDATE.OBJ_%s", s.name),
	)
	return s
}

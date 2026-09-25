package natsacl

import "slices"

// Builder
type Builder struct {
	pub map[string]any
	sub map[string]any
}

// NewBuilder returns a new NATS ACL builder
func NewBuilder() *Builder {
	return &Builder{
		pub: make(map[string]any),
		sub: make(map[string]any),
	}
}

// Sub returns the list of subscribe permissions
func (b *Builder) Sub() []string {
	keys := make([]string, 0, len(b.sub))
	for k := range b.sub {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Pub returns the list of publish permissions
func (b *Builder) Pub() []string {
	keys := make([]string, 0, len(b.pub))
	for k := range b.pub {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// AllowPub adds the provided subjects
func (b *Builder) AllowPub(subjects ...string) {
	for _, s := range subjects {
		b.pub[s] = nil
	}
}

// AllowSub adds the provided subjects
func (b *Builder) AllowSub(subjects ...string) {
	for _, s := range subjects {
		b.sub[s] = nil
	}
}

func (b *Builder) inbox() {
	b.sub["_INBOX.>"] = nil
}

func (b *Builder) pubInbox() {
	b.pub["_INBOX.>"] = nil
}

func (b *Builder) info() {
	b.pub["$JS.API.INFO"] = nil
	b.inbox()
}

func (b *Builder) names() {
	b.pub["$JS.API.STREAM.NAMES"] = nil
	b.inbox()

}

func (b *Builder) list() {
	b.pub["$JS.API.STREAM.LIST"] = nil
	b.inbox()
}

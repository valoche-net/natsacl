package natsacl

// Permissions represent NATS publish and subscribe permissions
type Permissions struct {
	Publish   []string
	Subscribe []string
}

// New returns empty NATS permissions
func New() *Permissions {
	return &Permissions{}
}

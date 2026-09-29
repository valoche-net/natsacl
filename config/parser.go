package config

import (
	"fmt"
	"io"

	"github.com/valoche-net/natsacl/v2"
	"go.yaml.in/yaml/v3"
)

type Permissions map[string][]Permission
type Permission map[string]any

// invalid formats the error depending in whether err is nil or not
func invalid(format string, value any, err error) error {
	if err != nil {
		return fmt.Errorf(format, err)
	}
	return fmt.Errorf(format, value)
}

func parseKV(builder *natsacl.Builder, perm Permission) error {
	name, err := mustString(perm["kv"])
	if err != nil {
		return err
	}

	kvbuilder := builder.KV(name)

	for k, v := range perm {
		switch k {
		case "kv":
		case "read":
			keys, err := getReadWriteKeys(v)
			if keys == nil {
				return invalid("invalid value %q", v, err)
			}
			kvbuilder.Read(*keys...)
		case "write":
			keys, err := getReadWriteKeys(v)
			if keys == nil {
				return invalid("invalid value %q", v, err)
			}
			kvbuilder.Write(*keys...)
		case "allow":
			allowed, err := getList(v)
			if allowed == nil {
				return invalid("expecting a list, got %q", v, err)
			}
			for _, allow := range *allowed {
				switch allow {
				case "*":
					kvbuilder.View().
						List().
						Create().
						Delete()
				case "list":
					kvbuilder.List()
				case "view":
					kvbuilder.View()
				case "create":
					kvbuilder.Create()
				case "delete":
					kvbuilder.Delete()
				default:
					return fmt.Errorf("K/V allowed %q incorect", allow)
				}
			}
		default:
			return fmt.Errorf("property %q is not valid for a K/V", k)
		}

	}
	return nil
}

func parseObj(builder *natsacl.Builder, perm Permission) error {
	name, err := mustString(perm["obj"])
	if err != nil {
		return err
	}

	objbuilder := builder.Obj(name)

	for k, v := range perm {
		switch k {
		case "obj":
		case "read":
			keys, err := getReadWriteKeys(v)
			if keys == nil {
				return invalid("invalid value %q", v, err)
			}
			objbuilder.Read(*keys...)
		case "write":
			keys, err := getReadWriteKeys(v)
			if keys == nil {
				return invalid("invalid value %q", v, err)
			}
			objbuilder.Write(*keys...)
		case "allow":
			allowed, err := getList(v)
			if err != nil {
				return invalid("invalid list %q", nil, err)
			}
			for _, allow := range *allowed {
				switch allow {
				case "*":
					objbuilder.View().
						Create().
						Delete().
						List().
						Seal()
				case "view":
					objbuilder.View()
				case "create":
					objbuilder.Create()
				case "delete":
					objbuilder.Delete()
				case "seal":
					objbuilder.Seal()
				case "list":
					objbuilder.List()
				default:
					return fmt.Errorf("Object Store allowed %q incorect", allow)
				}
			}
		default:
			return fmt.Errorf("property %q is not valid for an Object Store", k)
		}

	}
	return nil
}

func parseService(builder *natsacl.Builder, perm Permission) error {
	name, err := mustString(perm["service"])
	if err != nil {
		return err
	}
	subjects, err := getList(perm["subjects"])
	if err != nil {
		return fmt.Errorf("missing subjects for service %q: %v", name, err)
	}
	svcbuilder := builder.Service(name, *subjects...)
	for k, v := range perm {
		switch k {
		case "service":
		case "subjects":
		case "allow":
			allowed, err := getList(v)
			if err != nil {
				return fmt.Errorf("allow property invalid: %v", err)
			}
			for _, allow := range *allowed {
				switch allow {
				case "*":
					svcbuilder.Request().
						Respond().
						Monitor()
				case "request":
					svcbuilder.Request()
				case "respond":
					svcbuilder.Respond()
				case "monitor":
					svcbuilder.Monitor()
				default:
					return fmt.Errorf("Service allowed %q incorect", allow)
				}
			}
		default:
			return fmt.Errorf("property %q is not valid for a Service", k)
		}
	}
	return nil
}

func parseStream(builder *natsacl.Builder, perm Permission) error {
	name, err := mustString(perm["stream"])
	if err != nil {
		return err
	}
	streambuilder := builder.Stream(name)
	for k, v := range perm {
		switch k {
		case "stream":
		case "allow":
			allowed, err := getList(v)
			if err != nil {
				return fmt.Errorf("allow property invalid: %v", err)
			}
			for _, allow := range *allowed {
				switch allow {
				case "*":
					streambuilder.List().
						Info().
						Create().
						Delete().
						Read().
						Write().
						Purge()
				case "list":
					streambuilder.List()
				case "info":
					streambuilder.Info()
				case "create":
					streambuilder.Create()
				case "delete":
					streambuilder.Delete()
				case "read":
					streambuilder.Read()
				case "write":
					streambuilder.Write()
				case "purge":
					streambuilder.Purge()
				default:
					return fmt.Errorf("Stream allowed %q incorect", allow)
				}
			}
		default:
			return fmt.Errorf("property %q is not valid for a Stream", k)
		}
	}
	return nil
}

func parsePermission(builder *natsacl.Builder, perm Permission) (result error) {
	foundType := false
	isUnique := func() (bool, error) {
		if foundType {
			return false, fmt.Errorf("type must be only one of kv, obj, service or stream")
		}
		foundType = true
		return true, nil
	}
	for k := range perm {
		switch k {
		case "kv":
			if u, err := isUnique(); !u {
				return err
			}
			result = parseKV(builder, perm)
		case "obj":
			if u, err := isUnique(); !u {
				return err
			}
			result = parseObj(builder, perm)
		case "service":
			if u, err := isUnique(); !u {
				return err
			}
			result = parseService(builder, perm)
		case "stream":
			if u, err := isUnique(); !u {
				return err
			}
			result = parseStream(builder, perm)
		case "allow":
		case "read":
		case "write":
		case "subjects":
		default:
			return fmt.Errorf("unknown token %q", k)
		}
	}
	return
}

func parsePermissions(name string, perms []Permission) (*natsacl.Permissions, error) {
	builder := natsacl.NewBuilder()
	for _, p := range perms {
		if err := parsePermission(builder, p); err != nil {
			return nil, fmt.Errorf("parse permission error for %q: %q", name, err)
		}
	}
	return &natsacl.Permissions{
		Publish:   builder.Pub(),
		Subscribe: builder.Sub(),
	}, nil
}

func Parse(r io.Reader) (map[string]*natsacl.Permissions, error) {
	var perms Permissions
	if err := yaml.NewDecoder(r).Decode(&perms); err != nil {
		return nil, fmt.Errorf("cannot parse file: %q", err)
	}
	permissions := make(map[string]*natsacl.Permissions)
	for k, p := range perms {
		sperms, err := parsePermissions(k, p)
		if err != nil {
			return nil, err
		}
		permissions[k] = sperms
	}
	return permissions, nil
}

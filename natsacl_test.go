package natsacl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func runNATSserver(t *testing.T, users ...*server.User) *server.Server {
	t.Helper()

	users = append(users, &server.User{
		Username: "admin",
		Password: "admin",
		Permissions: &server.Permissions{
			Publish: &server.SubjectPermission{
				Allow: []string{">"},
			},
			Subscribe: &server.SubjectPermission{
				Allow: []string{">"},
			},
		},
	})
	opts := &server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      -1,
		Users:     users,
	}
	s, err := server.NewServer(opts)
	require.NoError(t, err)
	go s.Start()

	require.True(t, s.ReadyForConnections(5*time.Second), "NATS server failed to start")

	t.Cleanup(func() {
		s.Shutdown()
		s.WaitForShutdown()
	})
	return s
}

func requireNoErrors(t *testing.T, errCh <-chan error) {
	t.Helper()

	select {
	case err := <-errCh:
		t.Fatalf("expected no error, got %q", err)
	case <-time.After(250 * time.Millisecond):
		return
	}
}

func requirePermissionViolation(t *testing.T, errCh <-chan error, subject string) {
	t.Helper()

	select {
	case err := <-errCh:
		require.Contains(t, err.Error(), "Permissions Violation")
		require.Contains(t, err.Error(), fmt.Sprintf(`"%s"`, subject))
	case <-time.After(250 * time.Millisecond):
		t.Fatalf("expected permission violation for %q", subject)
	}
}

func runNATSserverWithPerms(t *testing.T, perms ...*Builder) *server.Server {
	t.Helper()

	var users []*server.User
	for i, perm := range perms {
		username := "testuser"
		if len(perms) > 1 {
			username = fmt.Sprintf("testuser%d", i)
		}

		users = append(users, &server.User{
			Username: username,
			Password: "testuser",
			Permissions: &server.Permissions{
				Publish: &server.SubjectPermission{
					Allow: perm.Pub(),
				},
				Subscribe: &server.SubjectPermission{
					Allow: perm.Sub(),
				},
			},
		})
	}

	return runNATSserver(t, users...)
}

func connectAdmin(t *testing.T, s *server.Server) *nats.Conn {
	t.Helper()

	nc, err := nats.Connect(
		s.ClientURL(),
		nats.UserInfo("admin", "admin"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { nc.Close() })

	return nc
}

func connectTestUser(t *testing.T, s *server.Server) (*nats.Conn, jetstream.JetStream, <-chan error) {
	t.Helper()

	/*
		Create object store and kv store for testing

		kv : key0 -> key4
		obj: file0 -> file4
	*/
	ncAdmin := connectAdmin(t, s)
	jsAdmin, err := jetstream.New(ncAdmin)
	require.NoError(t, err, "cannot get a Jetstream context as admin")

	kv, err := jsAdmin.CreateKeyValue(t.Context(), jetstream.KeyValueConfig{
		Bucket:      "kv",
		Description: "test kv",
	})
	require.NoError(t, err, "cannot create test kv")
	for k := range 5 {
		_, err := kv.PutString(t.Context(), fmt.Sprintf("key%d", k), fmt.Sprintf("value %d", k))
		require.NoError(t, err, "cannot create key")
	}

	obj, err := jsAdmin.CreateObjectStore(t.Context(), jetstream.ObjectStoreConfig{
		Bucket:      "obj",
		Description: "test obj",
	})
	require.NoError(t, err, "cannot create test object store")
	for k := range 5 {
		_, err := obj.PutString(t.Context(), fmt.Sprintf("file%d", k), fmt.Sprintf("content %d", k))
		require.NoError(t, err, "cannot create entry")
	}

	errCh := make(chan error, 10)

	nc, err := nats.Connect(
		s.ClientURL(),
		nats.UserInfo("testuser", "testuser"),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			t.Log("registered NATS error", err)
			errCh <- err
		}),
	)
	require.NoError(t, err, "cannot connect to NATS as testuser")
	t.Cleanup(func() { nc.Close() })

	js, err := jetstream.New(nc)
	require.NoError(t, err, "cannot get a Jetstream context as testuser")

	return nc, js, errCh
}

// fastContext is used to prevent JetStream 5 secondes timeout for testing operations
func fastContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)

	t.Cleanup(cancel)
	return ctx
}

func clearErrors(t *testing.T, nc *nats.Conn, errCh <-chan error) {
	t.Helper()

	require.NoError(t, nc.Flush())

	for {
		select {
		case <-errCh:
			// drain
		default:
			return
		}
	}
}

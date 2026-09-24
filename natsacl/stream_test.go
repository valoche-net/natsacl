package natsacl

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func TestStreamCreate(t *testing.T) {
	perms := NewBuilder().Stream("stream").Create().Build()

	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)

	_, err := js.CreateStream(t.Context(), jetstream.StreamConfig{
		Name:        "stream",
		Description: "Test stream",
		Subjects:    []string{"dummysub.>"},
	})
	require.NoError(t, err)

	_, err = js.CreateStream(deniedContext(t), jetstream.StreamConfig{
		Name:        "stream2",
		Description: "Test stream which will fail",
		Subjects:    []string{"dummysub2.>"},
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)

	_, err = js.Publish(deniedContext(t), "dummy.test", []byte("payload"))
	require.ErrorIs(t, err, context.DeadlineExceeded)

	clearErrors(t, nc, errCh)

	err = nc.Publish("this.is.denied", []byte("dummy"))
	require.NoError(t, err)
	require.NoError(t, nc.Flush())

	requirePermissionViolation(t, errCh, "this.is.denied")
}

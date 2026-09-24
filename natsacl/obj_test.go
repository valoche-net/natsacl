package natsacl

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func TestOBJCreate(t *testing.T) {
	perms := NewBuilder().Obj("myobj").Create().Build()

	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)
	var obj jetstream.ObjectStore
	var err error
	t.Run("can create requested obj store", func(t *testing.T) {
		obj, err = js.CreateObjectStore(t.Context(), jetstream.ObjectStoreConfig{
			Bucket:      "myobj",
			Description: "A test obj store",
		})
		require.NoError(t, err)
	})

	t.Run("cannot create another obj store", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := js.CreateObjectStore(deniedContext(t), jetstream.ObjectStoreConfig{
			Bucket:      "random",
			Description: "A random obj store",
		})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		requirePermissionViolation(t, errCh, "$JS.API.STREAM.CREATE.OBJ_random")
	})

	t.Run("cannot create entry in object store", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.PutString(deniedContext(t), "entry", "entry")
		b64entry := base64.StdEncoding.EncodeToString([]byte("entry"))
		require.ErrorIs(t, err, context.DeadlineExceeded)
		requirePermissionViolation(t, errCh, fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_myobj.$O.myobj.M.%s", b64entry))
	})
}

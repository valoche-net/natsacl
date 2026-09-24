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
		requireNoErrors(t, errCh)
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

func TestObjRead(t *testing.T) {
	perms := NewBuilder().Obj("obj").Read("file1", "file2").Write("testme").Build()

	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)

	obj, err := js.ObjectStore(t.Context(), "obj")
	require.NoError(t, err)
	t.Run("can read authorized entries", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.GetString(t.Context(), "file1")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
		_, err = obj.GetString(t.Context(), "file2")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
	})

	t.Run("can't read other entries", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.GetString(deniedContext(t), "file3")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		b64file := base64.StdEncoding.EncodeToString([]byte("file3"))
		requirePermissionViolation(t, errCh, fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_obj.$O.obj.M.%s", b64file))
	})

	t.Run("can modify data for allowed key", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.PutString(t.Context(), "testme", "testme")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
	})

	t.Run("can't modify data for not allowed key", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.PutString(deniedContext(t), "testme2", "testme")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		b64file := base64.StdEncoding.EncodeToString([]byte("testme2"))
		requirePermissionViolation(t, errCh, fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_obj.$O.obj.M.%s", b64file))
	})
}

func TestObjList(t *testing.T) {
	perms := NewBuilder().Obj("*").List().Build()
	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)

	t.Run("can list the object stores", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		names := js.ObjectStoreNames(deniedContext(t))
		require.NoError(t, names.Error())
		requireNoErrors(t, errCh)
	})
}

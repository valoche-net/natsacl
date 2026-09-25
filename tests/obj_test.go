package natsacl_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
	"github.com/valoche-net/natsacl/v2"
)

func TestOBJCreate(t *testing.T) {
	perms := natsacl.NewBuilder().Obj("myobj").Create().Build()

	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)
	var obj jetstream.ObjectStore
	var err error
	t.Run("can create requested obj store", func(t *testing.T) {
		obj, err = js.CreateObjectStore(fastContext(t), jetstream.ObjectStoreConfig{
			Bucket:      "myobj",
			Description: "A test obj store",
		})
		require.NoError(t, err)
		requireNoErrors(t, errCh)
	})

	t.Run("cannot create another obj store", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := js.CreateObjectStore(fastContext(t), jetstream.ObjectStoreConfig{
			Bucket:      "random",
			Description: "A random obj store",
		})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		requirePermissionViolation(t, errCh, "$JS.API.STREAM.CREATE.OBJ_random")
	})

	t.Run("cannot create entry in object store", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.PutString(fastContext(t), "entry", "entry")
		b64entry := base64.StdEncoding.EncodeToString([]byte("entry"))
		require.ErrorIs(t, err, context.DeadlineExceeded)
		requirePermissionViolation(t, errCh, fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_myobj.$O.myobj.M.%s", b64entry))
	})
}

func TestObjReadAndWrite(t *testing.T) {
	perms := natsacl.NewBuilder().Obj("obj").Read("file1", "file2").Write("testme").Build()

	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)

	obj, err := js.ObjectStore(fastContext(t), "obj")
	require.NoError(t, err)
	t.Run("can read authorized entries", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.GetString(fastContext(t), "file1")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
		_, err = obj.GetString(fastContext(t), "file2")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
	})

	t.Run("can't read other entries", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.GetString(fastContext(t), "file3")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		b64file := base64.StdEncoding.EncodeToString([]byte("file3"))
		requirePermissionViolation(t, errCh, fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_obj.$O.obj.M.%s", b64file))
	})

	t.Run("can modify data for allowed key", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.PutString(fastContext(t), "testme", "testme")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
	})

	t.Run("can't modify data for not allowed key", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		_, err := obj.PutString(fastContext(t), "testme2", "testme")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		b64file := base64.StdEncoding.EncodeToString([]byte("testme2"))
		requirePermissionViolation(t, errCh, fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_obj.$O.obj.M.%s", b64file))
	})
}

func TestObjList(t *testing.T) {
	perms := natsacl.NewBuilder().Obj("obj").List().Build()
	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)

	t.Run("can list the object stores", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		names := js.ObjectStoreNames(fastContext(t))
		require.NoError(t, names.Error())
		requireNoErrors(t, errCh)
		found := false
		for n := range names.Name() {
			if n == "obj" {
				found = true
			}
		}
		require.EqualValues(t, true, found, `could not find expected object store "obj"`)
	})
}

func TestObjAll(t *testing.T) {
	perms := natsacl.NewBuilder().Obj("*").All().Build()
	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)

	t.Run("can create an object store", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		obj, err := js.CreateOrUpdateObjectStore(fastContext(t), jetstream.ObjectStoreConfig{
			Bucket: "bigstore",
		})
		require.NoError(t, err)
		requireNoErrors(t, errCh)

		_, err = obj.PutString(fastContext(t), "file1", "data1")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
	})

	t.Run("can read and write any object store", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		obj1, err := js.ObjectStore(fastContext(t), "bigstore")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
		_, err = obj1.PutString(fastContext(t), "newdata", "newdata")
		require.NoError(t, err)
		requireNoErrors(t, errCh)

		obj2, err := js.ObjectStore(fastContext(t), "obj")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
		_, err = obj2.PutString(fastContext(t), "newdata", "newdata")
		require.NoError(t, err)
		requireNoErrors(t, errCh)
	})
}

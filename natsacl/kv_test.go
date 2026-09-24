package natsacl

import (
	"context"
	"fmt"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func TestKVCreate(t *testing.T) {
	perms := NewBuilder().KV("mykv").Create().Build()

	s := runNATSserverWithPerms(t, perms)
	nc, js, errCh := connectTestUser(t, s)

	t.Run("can create requested KV", func(t *testing.T) {
		kv, err := js.CreateKeyValue(t.Context(), jetstream.KeyValueConfig{
			Bucket:      "mykv",
			Description: "A test kv",
		})
		require.NoError(t, err)
		clearErrors(t, nc, errCh)
		_, err = kv.Create(deniedContext(t), "akey", []byte("a value"))
		require.Error(t, err)
		requirePermissionViolation(t, errCh, "$KV.mykv.akey")

	})
	t.Run("cannot create another KV", func(t *testing.T) {
		_, err := js.CreateKeyValue(t.Context(), jetstream.KeyValueConfig{
			Bucket:      "mykv",
			Description: "A test kv",
		})
		require.NoError(t, err)

	})

	t.Run("cannot list all the stores", func(t *testing.T) {
		clearErrors(t, nc, errCh)
		stores := js.KeyValueStoreNames(t.Context())
		require.NoError(t, stores.Error())
		requirePermissionViolation(t, errCh, "$JS.API.STREAM.NAMES")
	})
}

func TestKVReadAnyKey(t *testing.T) {
	perms := NewBuilder().KV("kv").Read().Build()

	s := runNATSserverWithPerms(t, perms)
	_, js, errCh := connectTestUser(t, s)

	kv, err := js.KeyValue(t.Context(), "kv")
	require.NoError(t, err)

	t.Run("get all the keys in the store", func(t *testing.T) {
		for i := range 5 {
			_, err := kv.Get(t.Context(), fmt.Sprintf("key%d", i))
			require.NoError(t, err)
		}
	})

	t.Run("cannot update a key", func(t *testing.T) {
		_, err = kv.PutString(deniedContext(t), "key1", "new value")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		requirePermissionViolation(t, errCh, "$KV.kv.key1")
	})

	t.Run("cannot create a new key", func(t *testing.T) {
		_, err = kv.PutString(deniedContext(t), "akey", "avalue")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		requirePermissionViolation(t, errCh, "$KV.kv.akey")
	})
}

func TestKVReadKey(t *testing.T) {
	perms := NewBuilder().KV("kv").Read("key1").Build()

	s := runNATSserverWithPerms(t, perms)
	_, js, errCh := connectTestUser(t, s)

	kv, err := js.KeyValue(t.Context(), "kv")
	require.NoError(t, err)

	t.Run("read authorized key", func(t *testing.T) {
		_, err := kv.Get(t.Context(), "key1")
		require.NoError(t, err)

	})

	t.Run("cannot read unauthorized key", func(t *testing.T) {
		_, err := kv.Get(deniedContext(t), "key2")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		requirePermissionViolation(t, errCh, "$JS.API.DIRECT.GET.KV_kv.$KV.kv.key2")
	})
}

func TestKVWriteAnyKey(t *testing.T) {
	perms := NewBuilder().KV("kv").Write().Build()
	s := runNATSserverWithPerms(t, perms)
	_, js, errCh := connectTestUser(t, s)

	kv, err := js.KeyValue(t.Context(), "kv")
	require.NoError(t, err)

	t.Run("write a new key", func(t *testing.T) {
		_, err := kv.PutString(t.Context(), "newkey", "newvalue")
		require.NoError(t, err)
	})

	t.Run("update an existing key", func(t *testing.T) {
		_, err := kv.PutString(t.Context(), "key1", "new value 1")
		require.NoError(t, err)
	})

	t.Run("cannot read a key", func(t *testing.T) {
		_, err := kv.Get(deniedContext(t), "key1")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		requirePermissionViolation(t, errCh, "$JS.API.DIRECT.GET.KV_kv.$KV.kv.key1")
	})

}

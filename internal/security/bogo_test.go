package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBogo(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		bogo := NewBogo(8, time.Minute)
		bogo.Wait()

		result := bogo.Compare()
		require.True(t, result)
	})

	t.Run("invalid number and timeout", func(t *testing.T) {
		bogo := NewBogo(0, time.Hour)
		bogo.Wait()

		result := bogo.Compare()
		require.True(t, result)
	})

	t.Run("timeout", func(t *testing.T) {
		bogo := NewBogo(1024, time.Second)
		bogo.Wait()

		result := bogo.Compare()
		require.False(t, result)
	})

	t.Run("cancel", func(t *testing.T) {
		bogo := NewBogo(1024, time.Minute)
		go func() {
			time.Sleep(time.Second)
			bogo.Stop()
		}()
		bogo.Wait()

		result := bogo.Compare()
		require.False(t, result)
	})
}

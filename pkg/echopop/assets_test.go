package echopop_test

import (
	"embed"
	"regexp"
	"testing"

	"github.com/invopop/client.go/pkg/echopop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/assets
var testAssets embed.FS

func TestAssetPath(t *testing.T) {
	// Matches "/<path>?v=<8 hex chars>".
	versioned := regexp.MustCompile(`^testdata/assets/app\.js\?v=[0-9a-f]{8}$`)

	t.Run("adds a version hash", func(t *testing.T) {
		p := echopop.AssetPath(testAssets, "testdata", "assets", "app.js")
		assert.Regexp(t, versioned, p[1:], "should be prefixed with a slash")
		assert.Equal(t, byte('/'), p[0])
	})

	t.Run("is stable across calls", func(t *testing.T) {
		first := echopop.AssetPath(testAssets, "testdata", "assets", "app.js")
		second := echopop.AssetPath(testAssets, "testdata", "assets", "app.js")
		assert.Equal(t, first, second, "the cache should return the same value")
	})

	t.Run("differs between files", func(t *testing.T) {
		js := echopop.AssetPath(testAssets, "testdata", "assets", "app.js")
		css := echopop.AssetPath(testAssets, "testdata", "assets", "style.css")
		require.NotEqual(t, js, css)
		assert.Contains(t, css, "style.css")
	})

	t.Run("falls back to the plain path when the file is missing", func(t *testing.T) {
		p := echopop.AssetPath(testAssets, "testdata", "assets", "missing.js")
		assert.Equal(t, "testdata/assets/missing.js", p, "should not be versioned")
		assert.NotContains(t, p, "?v=")
	})

	t.Run("joins the path segments", func(t *testing.T) {
		p := echopop.AssetPath(testAssets, "testdata/assets", "app.js")
		assert.Contains(t, p, "/testdata/assets/app.js?v=")
	})
}

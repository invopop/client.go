package echopop

import (
	"crypto/md5"
	"embed"
	"fmt"
	"io"
	"path"
)

var assetVersionCache = map[string]string{}

const (
	assetQueryFormat = "/%s?v=%s"
)

// AssetPath will find the file inside the embedded filesystem
// and attempt to add a version hash in the query parameter so that
// when the asset is loaded it will always provide the latest version.
//
// For example, to load a JS file inside a Templ component:
//
//	<script src={ echopop.AssetPath(assets.Content, "scripts", "app.js") }></script>
//
// Where `assets.Content` is the source of the file and `"scripts", "app.js"`
// identify the file's location. Output assumes that sources are from the root,
// for example, the above method might produce:
//
//	<script src="/scripts/app.js?12345678"></script>
//
// A simple version cache is used, and will only be renewed upon reloading
// the application. Paths must be unique for this to work correctly.
func AssetPath(content embed.FS, file ...string) string {
	p := path.Join(file...)
	if v, ok := assetVersionCache[p]; ok {
		return v
	}
	f, err := content.Open(p)
	if err != nil {
		return p
	}
	defer f.Close() //nolint:errcheck

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return p
	}
	v := fmt.Sprintf("%x", h.Sum(nil))[0:8]
	v = assetQueryPath(p, v)
	assetVersionCache[p] = v
	return v
}

func assetQueryPath(p, v string) string {
	return fmt.Sprintf(assetQueryFormat, p, v)
}

package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"strings"
)

type File struct {
	RelPath string
	URL     string
	Bytes   []byte
}

type StaticAssets struct {
	StylesheetURL   string
	ClientRenderURL string
}

func FingerprintPath(rel string, data []byte) string {
	cleaned := path.Clean(strings.TrimPrefix(rel, "/"))
	dir, file := path.Split(cleaned)
	ext := path.Ext(file)
	base := strings.TrimSuffix(file, ext)
	sum := sha256.Sum256(data)
	fingerprint := hex.EncodeToString(sum[:])[:8]
	name := base + "." + fingerprint + ext
	if dir == "" {
		return name
	}
	return path.Join(strings.TrimSuffix(dir, "/"), name)
}

func NewHashedFile(rel string, data []byte) File {
	outRel := path.Join("assets", FingerprintPath(rel, data))
	return File{
		RelPath: outRel,
		URL:     "/" + outRel,
		Bytes:   data,
	}
}

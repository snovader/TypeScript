package pnp

import (
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs"
)

func InitPnpApi(fs vfs.FS, filePath string) *PnpApi {
	pnpApi := &PnpApi{fs: fs, url: filePath}

	manifestData, err := pnpApi.findClosestPnpManifest()
	if err == nil {
		pnpApi.manifest = manifestData
		pnpApi.scopes = &pnpScopes{byPath: map[string]*PnpApi{manifestData.dirPath: pnpApi}}
		return pnpApi
	}

	return nil
}

type pnpScopes struct {
	mu     sync.Mutex
	byPath map[string]*PnpApi
}

// ForPath selects the closest PnP installation to an importing file or project.
// A language server can contain several independent installations, even when
// the editor opens only their common parent directory.
func (p *PnpApi) ForPath(filePath string) *PnpApi {
	if p.scopes == nil {
		return p
	}
	directoryPath := tspath.GetNormalizedAbsolutePath(filePath, p.manifest.dirPath)
	p.scopes.mu.Lock()
	defer p.scopes.mu.Unlock()

	var visited []string
	var scoped *PnpApi
	for {
		if cached, ok := p.scopes.byPath[directoryPath]; ok {
			scoped = cached
			break
		}
		visited = append(visited, directoryPath)
		if p.fs.FileExists(tspath.CombinePaths(directoryPath, ".pnp.cjs")) {
			manifest, err := parseManifestFromPath(p.fs, directoryPath)
			if err == nil {
				scoped = &PnpApi{fs: p.fs, url: directoryPath, manifest: manifest, scopes: p.scopes}
			}
			break
		}
		if tspath.IsDiskPathRoot(directoryPath) {
			break
		}
		directoryPath = tspath.GetDirectoryPath(directoryPath)
	}
	if scoped == nil {
		scoped = p
	}
	for _, path := range visited {
		p.scopes.byPath[path] = scoped
	}
	return scoped
}

func IsPnpLoaderFile(path string) bool {
	return strings.HasSuffix(path, ".pnp.cjs")
}

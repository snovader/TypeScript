package pnp

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
)

const nestedManifest = `{
  "dependencyTreeRoots": [{"name": "project", "reference": "workspace:."}],
  "enableTopLevelFallback": false,
  "fallbackPool": [],
  "fallbackExclusionList": [],
  "packageRegistryData": [
    ["project", [["workspace:.", {
      "packageLocation": "./",
      "packageDependencies": [["pkg", "npm:1.0.0"], ["@types/node", "npm:1.0.0"]]
    }]]],
    ["pkg", [["npm:1.0.0", {
      "packageLocation": "./.deps/pkg/",
      "packageDependencies": []
    }]]],
    ["@types/node", [["npm:1.0.0", {
      "packageLocation": "./.deps/@types/node/",
      "packageDependencies": []
    }]]]
  ]
}`

func TestNestedPnpResolution(t *testing.T) {
	t.Parallel()
	fs := vfstest.FromMap(map[string]string{
		"/repo/.pnp.cjs":                 "module.exports = {};",
		"/repo/.pnp.data.json":           nestedManifest,
		"/repo/sdk/.pnp.cjs":             "module.exports = {};",
		"/repo/sdk/.pnp.data.json":       nestedManifest,
		"/repo/.deps/pkg/index.d.ts":     "export const value: string;",
		"/repo/sdk/.deps/pkg/index.d.ts": "export const value: number;",
	}, true)
	api := InitPnpApi(fs, "/repo")
	for _, root := range []string{"/repo", "/repo/sdk", "/repo"} {
		resolved, err := api.ResolveToUnqualified("pkg", root+"/source/index.ts")
		if err != nil || resolved != root+"/.deps/pkg" {
			t.Fatalf("resolution from %s: got %q, %v", root, resolved, err)
		}
		typeRoots := api.GetPnpTypeRoots(root + "/source")
		if len(typeRoots) != 1 || typeRoots[0] != root+"/.deps/@types" {
			t.Fatalf("type roots from %s: got %v", root, typeRoots)
		}
		if !api.IsImportable(root+"/source/index.ts", root+"/.deps/pkg/index.d.ts") {
			t.Fatalf("dependency is not importable from %s", root)
		}
		if !api.IsInPnpModule(root+"/.deps/pkg/index.d.ts", root+"/source/index.ts") {
			t.Fatalf("dependency is not treated as an external module from %s", root)
		}
		if manifest := api.ForPath(root + "/source/index.ts").GetManifestPath(); manifest != root+"/.pnp.cjs" {
			t.Fatalf("manifest to watch from %s: got %q", root, manifest)
		}
	}
}

func TestNestedPnpRefresh(t *testing.T) {
	t.Parallel()
	fs := vfstest.FromMap(map[string]string{
		"/repo/.pnp.cjs":           "module.exports = {};",
		"/repo/.pnp.data.json":     nestedManifest,
		"/repo/sdk/.pnp.cjs":       "module.exports = {};",
		"/repo/sdk/.pnp.data.json": nestedManifest,
	}, true)
	api := InitPnpApi(fs, "/repo")
	issuer := "/repo/sdk/source/index.ts"
	api.ResolveToUnqualified("pkg", issuer)
	if err := fs.WriteFile("/repo/sdk/.pnp.data.json", strings.ReplaceAll(nestedManifest, ".deps/pkg/", ".deps/updated-pkg/")); err != nil {
		t.Fatal(err)
	}
	refreshed, err := api.RefreshManifest()
	if err != nil {
		t.Fatal(err)
	}
	resolved, resolveErr := refreshed.ResolveToUnqualified("pkg", issuer)
	if resolveErr != nil || resolved != "/repo/sdk/.deps/updated-pkg" {
		t.Fatalf("refreshed nested resolution: got %q, %v", resolved, resolveErr)
	}
	resolved, resolveErr = api.ResolveToUnqualified("pkg", issuer)
	if resolveErr != nil || resolved != "/repo/sdk/.deps/pkg" {
		t.Fatalf("previous snapshot's resolution changed: got %q, %v", resolved, resolveErr)
	}
}

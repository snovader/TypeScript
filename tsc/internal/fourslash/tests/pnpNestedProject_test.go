package fourslash_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/fourslash"
)

func TestPnpNestedProject(t *testing.T) {
	t.Parallel()
	const content = `
// @filename: /.pnp.cjs
module.exports = {};
// @filename: /.pnp.data.json
{
  "dependencyTreeRoots": [{"name":"root","reference":"workspace:."}],
  "enableTopLevelFallback": false,
  "fallbackPool": [],
  "fallbackExclusionList": [],
  "packageRegistryData": [
    ["root", [["workspace:.", {"packageLocation":"./","packageDependencies":[["pkg","npm:1.0.0"]]}]]],
    ["pkg", [["npm:1.0.0", {"packageLocation":"./.deps/pkg/","packageDependencies":[]}]]]
  ]
}
// @filename: /tsconfig.json
{"compilerOptions":{"module":"preserve","moduleResolution":"bundler","types":[]},"include":["src"]}
// @filename: /.deps/pkg/package.json
{"name":"pkg","version":"1.0.0","types":"index.d.ts"}
// @filename: /.deps/pkg/index.d.ts
export declare const value: string;
// @filename: /src/index.ts
import { value } from "pkg";
const rootValue: string = value;

// @filename: /sdk/.pnp.cjs
module.exports = {};
// @filename: /sdk/.pnp.data.json
{
  "dependencyTreeRoots": [{"name":"sdk","reference":"workspace:."}],
  "enableTopLevelFallback": false,
  "fallbackPool": [],
  "fallbackExclusionList": [],
  "packageRegistryData": [
    ["sdk", [["workspace:.", {"packageLocation":"./","packageDependencies":[["pkg","npm:2.0.0"]]}]]],
    ["pkg", [["npm:2.0.0", {"packageLocation":"./.deps/pkg/","packageDependencies":[]}]]]
  ]
}
// @filename: /sdk/tsconfig.json
{"compilerOptions":{"module":"preserve","moduleResolution":"bundler","types":[]},"include":["src"]}
// @filename: /sdk/.deps/pkg/package.json
{"name":"pkg","version":"2.0.0","types":"index.d.ts"}
// @filename: /sdk/.deps/pkg/index.d.ts
export declare const value: number;
// @filename: /sdk/src/index.ts
import { value } from "pkg";
const sdkValue: number = value;
`
	f, done := fourslash.NewFourslash(t, nil, content)
	defer done()
	f.GoToFile(t, "/src/index.ts")
	f.VerifyNumberOfErrorsInCurrentFile(t, 0)
	f.GoToFile(t, "/sdk/src/index.ts")
	f.VerifyNumberOfErrorsInCurrentFile(t, 0)
}

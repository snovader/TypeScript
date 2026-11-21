package project

import (
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/internal/pnp"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs"
)

// TsGoLintConfigRegistry owns a single-run config-discovery snapshot.
// It keeps compiler project internals behind the linter integration boundary.
type TsGoLintConfigRegistry struct {
	builder     *configFileRegistryBuilder
	directory   tspath.RootedDirectoryPath
	sensitivity tspath.CaseSensitivity
}

func NewTsGoLintConfigRegistry(fs vfs.FS, directory string) *TsGoLintConfigRegistry {
	cwd := tspath.RootedDirectoryPathFromAbsolute(directory)
	source := layerOverlayFileSystem(fs, nil, lsproto.PositionEncodingKindUTF16)
	snapshot := newSnapshotFSBuilderFromSource(source, nil, nil, nil)
	builder := newConfigFileRegistryBuilder(false, snapshot, func(tspath.PathKey) bool { return false },
		&ConfigFileRegistry{}, NewExtendedConfigCache(), 0,
		&SessionOptions{CurrentDirectory: cwd}, pnp.InitPnpApi(fs, directory), "", nil)
	return &TsGoLintConfigRegistry{builder: builder, directory: cwd, sensitivity: fs.CaseSensitivity()}
}

func (r *TsGoLintConfigRegistry) ComputeConfigFileName(file string, skip bool) string {
	return r.builder.computeConfigFileName(tspath.ToRootedFilePath(file, r.directory), skip, nil).AsString()
}

func (r *TsGoLintConfigRegistry) AcquireConfig(config string, file string) *tsoptions.ParsedCommandLine {
	name := tspath.ToRootedFilePath(config, r.directory)
	fileName := tspath.ToRootedFilePath(file, r.directory)
	return r.builder.findOrAcquireConfigForFile(name, r.sensitivity.PathKey(name.AsPath()),
		r.sensitivity.PathKey(fileName.AsPath()), projectLoadKindCreate, nil)
}

func (r *TsGoLintConfigRegistry) AncestorConfig(file string, config string) string {
	name := tspath.ToRootedFilePath(file, r.directory)
	return r.builder.getAncestorConfigFileName(name, r.sensitivity.PathKey(name.AsPath()),
		tspath.ToRootedFilePath(config, r.directory), nil).AsString()
}

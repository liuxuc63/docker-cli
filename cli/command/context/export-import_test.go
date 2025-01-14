package context

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/streams"
	"gotest.tools/v3/assert"
)

func TestExportImportWithFile(t *testing.T) {
	contextDir, err := ioutil.TempDir("", t.Name()+"context")
	assert.NilError(t, err)
	defer os.RemoveAll(contextDir)
	contextFile := filepath.Join(contextDir, "exported")
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContext(t, cli)
	cli.ErrBuffer().Reset()
	assert.NilError(t, RunExport(cli, &ExportOptions{
		ContextName: "test",
		Dest:        contextFile,
	}))
	assert.Equal(t, cli.ErrBuffer().String(), fmt.Sprintf("Written file %q\n", contextFile))
	cli.OutBuffer().Reset()
	cli.ErrBuffer().Reset()
	assert.NilError(t, RunImport(cli, "test2", contextFile))
	context1, err := cli.ContextStore().GetMetadata("test")
	assert.NilError(t, err)
	context2, err := cli.ContextStore().GetMetadata("test2")
	assert.NilError(t, err)
	assert.DeepEqual(t, context1.Endpoints, context2.Endpoints)
	assert.DeepEqual(t, context1.Metadata, context2.Metadata)
	assert.Equal(t, "test", context1.Name)
	assert.Equal(t, "test2", context2.Name)

	assert.Equal(t, "test2\n", cli.OutBuffer().String())
	assert.Equal(t, "Successfully imported context \"test2\"\n", cli.ErrBuffer().String())
}

func TestExportImportPipe(t *testing.T) {
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	createTestContext(t, cli)
	cli.ErrBuffer().Reset()
	cli.OutBuffer().Reset()
	assert.NilError(t, RunExport(cli, &ExportOptions{
		ContextName: "test",
		Dest:        "-",
	}))
	assert.Equal(t, cli.ErrBuffer().String(), "")
	cli.SetIn(streams.NewIn(ioutil.NopCloser(bytes.NewBuffer(cli.OutBuffer().Bytes()))))
	cli.OutBuffer().Reset()
	cli.ErrBuffer().Reset()
	assert.NilError(t, RunImport(cli, "test2", "-"))
	context1, err := cli.ContextStore().GetMetadata("test")
	assert.NilError(t, err)
	context2, err := cli.ContextStore().GetMetadata("test2")
	assert.NilError(t, err)
	assert.DeepEqual(t, context1.Endpoints, context2.Endpoints)
	assert.DeepEqual(t, context1.Metadata, context2.Metadata)
	assert.Equal(t, "test", context1.Name)
	assert.Equal(t, "test2", context2.Name)

	assert.Equal(t, "test2\n", cli.OutBuffer().String())
	assert.Equal(t, "Successfully imported context \"test2\"\n", cli.ErrBuffer().String())
}

func TestExportExistingFile(t *testing.T) {
	contextDir, err := ioutil.TempDir("", t.Name()+"context")
	assert.NilError(t, err)
	defer os.RemoveAll(contextDir)
	contextFile := filepath.Join(contextDir, "exported")
	cli, cleanup := makeFakeCli(t)
	defer cleanup()
	cli.ErrBuffer().Reset()
	assert.NilError(t, ioutil.WriteFile(contextFile, []byte{}, 0644))
	err = RunExport(cli, &ExportOptions{ContextName: "test", Dest: contextFile})
	assert.Assert(t, os.IsExist(err))
}

func createTestContext(t *testing.T, cli command.Cli) {
	t.Helper()

	err := RunCreate(cli, &CreateOptions{
		Name:   "test",
		Docker: map[string]string{},
	})
	assert.NilError(t, err)
}

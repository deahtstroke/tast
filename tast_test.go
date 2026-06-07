package tast_test

import (
	"os"
	"testing"

	"github.com/deahtstroke/tast"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func Test_ParseKeyValue(t *testing.T) {
	src, err := os.ReadFile("testdata/parse/test.toml")
	assert.NilError(t, err, "error reading from file parse/test.toml")

	doc, err := tast.ParseString(string(src))
	assert.NilError(t, err, "not expecting error, got: %v", err)

	kv := requireKeyValue(t, doc, "src")
	val := requireString(t, kv)

	assert.Equal(t, val, "output.lua")
}

func Test_ParseTable(t *testing.T) {
	src, err := os.ReadFile("testdata/parse/test.toml")
	assert.NilError(t, err, "error reading from file parse/test.toml")

	doc, err := tast.ParseBytes(src)
	assert.NilError(t, err, "not expecting error, got: %v", err)

	table := requireTable(t, doc, "foo")
	assert.Equal(t, table.Key(), "foo")
}

func Test_ParseKeyValueInTable(t *testing.T) {
	src, err := os.ReadFile("testdata/parse/test.toml")
	assert.NilError(t, err, "error reading from file parse/test.toml")

	doc, err := tast.ParseBytes(src)
	assert.NilError(t, err, "not expecting error, got: %v", err)

	table := requireTable(t, doc, "foo")
	kv := requireTableKeyValue(t, table, "bar")
	val := requireString(t, kv)

	assert.Equal(t, val, "Hello!")
}

func Test_ModifyExistingValueInTable(t *testing.T) {
	src, err := os.ReadFile("testdata/parse/test.toml")
	assert.NilError(t, err, "error reading from file parse/test.toml")

	doc, err := tast.ParseBytes(src)
	assert.NilError(t, err, "not expecting error, got: %v", err)

	table := requireTable(t, doc, "foo")
	kv := requireTableKeyValue(t, table, "bar")

	assert.NilError(t, kv.Set(1234))

	kv = requireTableKeyValue(t, table, "bar")
	val := requireInt(t, kv)
	assert.Equal(t, val, int64(1234))
}

func Test_DeleteExistingKeyInTable(t *testing.T) {
	src, err := os.ReadFile("testdata/parse/test.toml")
	assert.NilError(t, err, "error reading from file parse/test.toml")

	doc, err := tast.ParseBytes(src)
	assert.NilError(t, err, "not expecting error, got: %v", err)

	table := requireTable(t, doc, "foo")
	assert.Equal(t, table.Delete("bar"), true)

	_, ok := table.FindKey("bar")
	assert.Equal(t, ok, false)
}

func Test_RoundTrip(t *testing.T) {
	src, err := os.ReadFile("testdata/roundtrip/test.toml")
	assert.NilError(t, err, "error reading from roundtrip/test.toml")

	doc, err := tast.ParseBytes(src)
	assert.NilError(t, err, "not expecting error, got: %v", err)

	s, err := doc.String()
	assert.NilError(t, err, "not expecting error, got: %v", err)

	assert.Assert(t, cmp.Contains(s, "# Database details"))
	assert.Assert(t, cmp.Contains(s, "port = 5432"))
	assert.Assert(t, cmp.Contains(s, "host = \"localhost\""))
	assert.Assert(t, cmp.Contains(s, "username = \"root\""))
	assert.Assert(t, cmp.Contains(s, "password = \"root\""))
}

func requireKeyValue(t *testing.T, doc *tast.Document, name string) *tast.KeyValueNode {
	t.Helper()
	kv, ok := doc.FindKey(name)
	assert.Assert(t, ok, "key %q not found", name)
	return kv
}

func requireTable(t *testing.T, doc *tast.Document, name string) *tast.TableNode {
	t.Helper()
	table, ok := doc.Table(name)
	assert.Assert(t, ok, "table %q not found", name)
	return table
}

func requireTableKeyValue(t *testing.T, table *tast.TableNode, key string) *tast.KeyValueNode {
	t.Helper()
	kv, ok := table.FindKey(key)
	assert.Assert(t, ok, "key %q not found", key)
	return kv
}

func requireString(t *testing.T, kv *tast.KeyValueNode) string {
	t.Helper()
	val, ok := kv.String()
	assert.Assert(t, ok, "expected string value, got %T", kv)
	return val
}

func requireInt(t *testing.T, kv *tast.KeyValueNode) any {
	t.Helper()
	val, ok := kv.Int()
	assert.Assert(t, ok, "expected integer value, got %T", kv)
	return val
}

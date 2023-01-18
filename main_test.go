package jsonnet_filer_lib

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/zeet-dev/jsonnet-filer-lib/internal/sh"
)

func Test_main_jsonnet(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	exitCode, err := sh.Run(context.Background(), "jsonnet", func(o *sh.RunOptions) {
		o.Args = []string{"./main.libsonnet"}
		o.Stdout = &out
		o.Stderr = &errOut
	})

	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "{ }\n", out.String())
	assert.Equal(t, "", errOut.String())
}

type ObjectMeta struct {
	Name string `json:"name"`
}

type File struct {
	ApiVersion       string     `json:"apiVersion"`
	Kind             string     `json:"kind"`
	Metadata         ObjectMeta `json:"metadata"`
	Content          any        `json:"content"`
	ContentEncoded   string     `json:"contentEncoded"`
	EncodingStrategy string     `json:"encodingStrategy"`
}

func Test_empty_file(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	exitCode, err := sh.Run(context.Background(), "jsonnet", func(o *sh.RunOptions) {
		o.Args = []string{
			"--exec",
			`
local jf = import "./main.libsonnet";
jf.File("foo")
`,
		}
		o.Stdout = &out
		o.Stderr = &errOut
	})

	expectedFile := File{
		ApiVersion: "jsonnet-filer.zeet.co/v1alpha1",
		Kind:       "File",
		Metadata: ObjectMeta{
			Name: "foo",
		},
		Content:          "",
		ContentEncoded:   `""`,
		EncodingStrategy: "yaml",
	}

	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	actualFile := File{}
	err = json.Unmarshal(out.Bytes(), &actualFile)
	require.NoError(t, err)
	assert.Equal(t, expectedFile, actualFile)
}

func Test_yaml_file(t *testing.T) {
	content := map[string]any{
		"foo": "bar",
		"fuz": []any{"item1", "item2"},
		"objhere": map[string]any{
			"inner": "v",
		},
	}
	contentJson, err := json.Marshal(&content)
	require.NoError(t, err)

	var out bytes.Buffer
	var errOut bytes.Buffer
	exitCode, err := sh.Run(context.Background(), "jsonnet", func(o *sh.RunOptions) {
		o.Args = []string{
			"--exec",
			`local jf = import "./main.libsonnet";
jf.File("foo",` + string(contentJson) + `)
`,
		}
		o.Stdout = &out
		o.Stderr = &errOut
	})

	contentYamlEncoded, err := yaml.Marshal(&content)

	expectedFile := File{
		ApiVersion: "jsonnet-filer.zeet.co/v1alpha1",
		Kind:       "File",
		Metadata: ObjectMeta{
			Name: "foo",
		},
		Content:          content,
		ContentEncoded:   string(contentYamlEncoded),
		EncodingStrategy: "yaml",
	}

	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	actualFile := File{}
	err = json.Unmarshal(out.Bytes(), &actualFile)
	require.NoError(t, err)

	// 2023-01-17 wes@zeet.co
	// jsonnet yaml marshalling _always_ quotes values
	// before we assert that the whole file matches, we need to massage the content returned
	// by unmarshalling it and remarshalling it to get a consistent value
	// https://github.com/google/go-jsonnet/issues/665
	var actualContent any
	err = yaml.Unmarshal([]byte(actualFile.ContentEncoded), &actualContent)
	require.NoError(t, err)
	actualContentRemarshalled, err := yaml.Marshal(actualContent)
	require.NoError(t, err)
	actualFile.ContentEncoded = string(actualContentRemarshalled)

	assert.Equal(t, expectedFile, actualFile)
}

func Test_json_file(t *testing.T) {
	content := map[string]any{
		"foo": "bar",
		"fuz": []any{"item1", "item2"},
		"objhere": map[string]any{
			"inner": "v",
		},
	}
	contentJsonBytes, err := json.MarshalIndent(&content, "", "  ")
	require.NoError(t, err)
	contentJson := string(contentJsonBytes)

	var out bytes.Buffer
	var errOut bytes.Buffer
	exitCode, err := sh.Run(context.Background(), "jsonnet", func(o *sh.RunOptions) {
		o.Args = []string{
			"--exec",
			`local jf = import "./main.libsonnet";
jf.File("foo",` + contentJson + `) + { encodingStrategy: "json" }
`,
		}
		o.Stdout = &out
		o.Stderr = &errOut
	})

	expectedFile := File{
		ApiVersion: "jsonnet-filer.zeet.co/v1alpha1",
		Kind:       "File",
		Metadata: ObjectMeta{
			Name: "foo",
		},
		Content:          content,
		ContentEncoded:   contentJson,
		EncodingStrategy: "json",
	}

	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	actualFile := File{}
	err = json.Unmarshal(out.Bytes(), &actualFile)
	require.NoError(t, err)
	assert.Equal(t, expectedFile, actualFile)
}

// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2019 Datadog, Inc.

package util

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreValue(t *testing.T) {
	testDir, err := ioutil.TempDir("", "fake-datadog-var-")
	require.Nil(t, err, fmt.Sprintf("%v", err))
	defer os.RemoveAll(testDir)
	mockConfig := config.Mock()
	mockConfig.Set("var_path", testDir)
	err = StoreValue("mykey", "myvalue")
	assert.Nil(t, err)
	assert.Equal(t, "myvalue", RetrieveValue("mykey"))
	assert.Equal(t, "", RetrieveValue("myotherkey"))
}

func TestStoreValueInvalidChar(t *testing.T) {
	testDir, err := ioutil.TempDir("", "fake-datadog-var-")
	require.Nil(t, err, fmt.Sprintf("%v", err))
	defer os.RemoveAll(testDir)
	mockConfig := config.Mock()
	mockConfig.Set("var_path", testDir)
	err = StoreValue("my:key", "myvalue")
	assert.Nil(t, err)
	assert.Equal(t, "myvalue", RetrieveValue("my:key"))

	expectPath := filepath.Join(testDir, "my_key")
	_, err = os.Stat(expectPath)
	require.Nil(t, err)
}

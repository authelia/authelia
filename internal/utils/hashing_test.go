package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/random"
)

func TestShouldHashString(t *testing.T) {
	input := "input"
	anotherInput := "another"

	sum := HashSHA256FromString(input)

	assert.Equal(t, "c96c6d5be8d08a12e7b5cdc1b207fa6b2430974c86803d8891675e76fd992c20", sum)

	anotherSum := HashSHA256FromString(anotherInput)

	assert.Equal(t, "ae448ac86c4e8e4dec645729708ef41873ae79c6dff84eff73360989487f08e5", anotherSum)
	assert.NotEqual(t, sum, anotherSum)

	randomInput := r.StringCustom(40, random.CharSetAlphaNumeric)
	randomSum := HashSHA256FromString(randomInput)

	assert.NotEqual(t, randomSum, sum)
	assert.NotEqual(t, randomSum, anotherSum)
}

func TestShouldHashPath(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "myfile"), []byte("output\n"), 0600)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "anotherfile"), []byte("another\n"), 0600)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "randomfile"), []byte(r.StringCustom(40, random.CharSetAlphaNumeric)+"\n"), 0600)
	assert.NoError(t, err)

	sum, err := HashSHA256FromPath(filepath.Join(dir, "myfile"))

	assert.NoError(t, err)
	assert.Equal(t, "9aff6ba4b042b9d09991a9fbf8c80ddbd2a9c433638339cd831bed955e39f106", sum)

	anotherSum, err := HashSHA256FromPath(filepath.Join(dir, "anotherfile"))

	assert.NoError(t, err)
	assert.Equal(t, "33a7b215065f2ee8635efb72620bc269a1efb889ba3026560334da7366742374", anotherSum)

	randomSum, err := HashSHA256FromPath(filepath.Join(dir, "randomfile"))

	assert.NoError(t, err)
	assert.NotEqual(t, randomSum, sum)
	assert.NotEqual(t, randomSum, anotherSum)

	sum, err = HashSHA256FromPath(filepath.Join(dir, "notafile"))
	assert.Equal(t, "", sum)

	errTxt := GetExpectedErrTxt("filenotfound")
	assert.EqualError(t, err, fmt.Sprintf(errTxt, filepath.Join(dir, "notafile")))
}

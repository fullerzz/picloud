package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAltSizes(t *testing.T) {
	srcPath := "testdata/baxter.jpg"
	createAltSizes(srcPath)

	// Check if 1/4 scale file generated
	_, err := os.Stat(fmt.Sprintf("%s-%d.jpeg", srcPath, 4))
	assert.NoError(t, err)

	// Check if 1/10 scale file generated
	_, err = os.Stat(fmt.Sprintf("%s-%d.jpeg", srcPath, 10))
	assert.NoError(t, err)
}

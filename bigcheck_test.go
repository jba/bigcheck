package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	*size = 100
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer, "a")
}

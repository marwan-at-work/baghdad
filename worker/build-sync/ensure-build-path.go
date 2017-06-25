package main

import (
	"fmt"
	"os"
)

func ensureBuildPath() {
	err := os.MkdirAll(ArtifactsPath, 0666)
	if err != nil {
		panic(fmt.Errorf("could not make pr dir: %v", err))
	}
}

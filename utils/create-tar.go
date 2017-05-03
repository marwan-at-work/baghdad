package utils

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/cli/command/image/build"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
)

// DefaultDockerfileName is used when no dockerfile was specified
const DefaultDockerfileName = "Dockerfile"

// CreateTar create a build context tar for the specified project and service name.
// copied from libcompose https://github.com/docker/libcompose/blob/master/docker/builder/builder.go
func CreateTar(contextDirectory, dockerfile string) (io.ReadCloser, error) {
	// This code was ripped off from docker/api/client/build.go
	dockerfileName := filepath.Join(contextDirectory, dockerfile)

	absContextDirectory, err := filepath.Abs(contextDirectory)
	if err != nil {
		return nil, err
	}

	filename := dockerfileName

	if dockerfile == "" {
		// No -f/--file was specified so use the default
		dockerfileName = DefaultDockerfileName
		filename = filepath.Join(absContextDirectory, dockerfileName)

		// Just to be nice ;-) look for 'dockerfile' too but only
		// use it if we found it, otherwise ignore this check
		if _, err = os.Lstat(filename); os.IsNotExist(err) {
			tmpFN := path.Join(absContextDirectory, strings.ToLower(dockerfileName))
			if _, err = os.Lstat(tmpFN); err == nil {
				dockerfileName = strings.ToLower(dockerfileName)
				filename = tmpFN
			}
		}
	}

	origDockerfile := dockerfileName // used for error msg
	if filename, err = filepath.Abs(filename); err != nil {
		return nil, err
	}

	// Now reset the dockerfileName to be relative to the build context
	dockerfileName, err = filepath.Rel(absContextDirectory, filename)
	if err != nil {
		return nil, err
	}

	// And canonicalize dockerfile name to a platform-independent one
	dockerfileName, err = archive.CanonicalTarNameForPath(dockerfileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot canonicalize dockerfile path %s: %v", dockerfileName, err)
	}

	if _, err = os.Lstat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cannot locate Dockerfile: %s", origDockerfile)
	}
	var includes = []string{"."}
	var excludes []string

	dockerIgnorePath := path.Join(contextDirectory, ".dockerignore")
	dockerIgnore, err := os.Open(dockerIgnorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		excludes = make([]string, 0)
	} else {
		excludes, err = dockerignore.ReadAll(dockerIgnore)
		if err != nil {
			return nil, err
		}
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed.  The deamon will remove them for us, if needed, after it
	// parses the Dockerfile.
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(dockerfileName, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", dockerfileName)
	}

	if err := build.ValidateContextDirectory(contextDirectory, excludes); err != nil {
		return nil, fmt.Errorf("error checking context is accessible: '%s', please check permissions and try again", err)
	}

	options := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	return archive.TarWithOptions(contextDirectory, options)
}

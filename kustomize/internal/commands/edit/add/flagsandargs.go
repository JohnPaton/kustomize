// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/internal/commands/util"
)

// flagsAndArgs encapsulates the options for add secret/configmap commands.
type flagsAndArgs struct {
	// Name of configMap/Secret (required)
	Name string
	// FileSources to derive the configMap/Secret from (optional)
	FileSources []string
	// LiteralSources to derive the configMap/Secret from (optional)
	LiteralSources []string
	// EnvFileSource to derive the configMap/Secret from (optional)
	// TODO: Rationalize this name with Generic.EnvSource
	EnvFileSource string
	// Type of secret to create
	Type string
}

// Validate validates required fields are set to support structured generation.
func (a *flagsAndArgs) Validate(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("name must be specified once")
	}
	a.Name = args[0]
	if len(a.EnvFileSource) == 0 && len(a.FileSources) == 0 && len(a.LiteralSources) == 0 {
		return fmt.Errorf("at least from-env-file, or from-file or from-literal must be set")
	}
	if len(a.EnvFileSource) > 0 && (len(a.FileSources) > 0 || len(a.LiteralSources) > 0) {
		return fmt.Errorf("from-env-file cannot be combined with from-file or from-literal")
	}
	// TODO: Should we check if the path exists? if it's valid, if it's within the same (sub-)directory?
	return nil
}

// ExpandFileSource normalizes a string list, possibly
// containing globs, into a validated, globless list.
// For example, this list:
//     some/path
//     some/dir/a*
//     bfile=some/dir/b*
// becomes:
//     some/path
//     some/dir/airplane
//     some/dir/ant
//     some/dir/apple
//     bfile=some/dir/banana
// i.e. everything is converted to a key=value pair,
// where the value is always a relative file path,
// and the key, if missing, is the same as the value.
// In the case where the key is explicitly declared,
// the globbing, if present, must have exactly one match.
func (a *flagsAndArgs) ExpandFileSource(fSys filesys.FileSystem) error {
	var results []string
	for _, pattern := range a.FileSources {
		var patterns []string
		key := ""
		// check if the pattern is in `--from-file=[key=]source` format
		// and if so split it to send only the file-pattern to glob function
		s := strings.Split(pattern, "=")
		if len(s) == 2 {
			patterns = append(patterns, s[1])
			key = s[0]
		} else {
			patterns = append(patterns, s[0])
		}
		result, err := util.GlobPatterns(fSys, patterns)
		if err != nil {
			return err
		}
		// if the format is `--from-file=[key=]source` accept only one result
		// and extend it with the `key=` prefix
		if key != "" {
			if len(result) != 1 {
				return fmt.Errorf(
					"'pattern '%s' catches files %v, should catch only one", pattern, result)
			}
			fileSource := fmt.Sprintf("%s=%s", key, result[0])
			results = append(results, fileSource)
		} else {
			results = append(results, result...)
		}
	}
	a.FileSources = results
	return nil
}

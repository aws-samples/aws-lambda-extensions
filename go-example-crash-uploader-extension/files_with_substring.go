// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func renameFilesWithSubstring(rootDirectory string, substring string, newName string) (int, error) {
	files, err := getFilesWithSubstring(rootDirectory, substring)
	if err != nil {
		return 0, err
	}

	for idx, file := range files {

		if idx == 0 {
			err = os.Rename(file, fmt.Sprintf("%s/%s", filepath.Dir(file), newName))
			if err != nil {
				return idx, err
			}
		} else {
			err = os.Rename(file, fmt.Sprintf("%s/%s(%d)", filepath.Dir(file), newName, idx))
			if err != nil {
				return idx, err
			}
		}
	}
	return len(files), nil
}

func getFilesWithSubstring(rootDirectory string, substring string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootDirectory, func(path string, info os.FileInfo, err error) error {
		substringFound := strings.Contains(path, substring)
		if substringFound {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

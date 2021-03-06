//-----------------------------------------------------------------------------
// Copyright (c) 2017 Oracle and/or its affiliates.  All rights reserved.
// This program is free software: you can modify it and/or redistribute it
// under the terms of:
//
// (i)  the Universal Permissive License v 1.0 or at your option, any
//      later version (http://oss.oracle.com/licenses/upl); and/or
//
// (ii) the Apache License v 2.0. (http://www.apache.org/licenses/LICENSE-2.0)
//-----------------------------------------------------------------------------

package main

import (
	"bytes"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/wercker/pkg/log"
)

func createWalker(cfg *Config, templateRoot, outputRoot string) func(string, os.FileInfo, error) error {
	return func(p string, info os.FileInfo, err error) error {
		relativePath := strings.TrimPrefix(p, templateRoot)
		if relativePath == "" {
			return nil
		}

		if info.Name() != "vendor" && info.Name() != "vendor.json" && strings.HasPrefix(relativePath, "/vendor") {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".swp") {
			return nil
		}

		replacedPath := replaceSentinels(relativePath)

		templatePath := path.Join(templateRoot, relativePath)

		tmpl, err := template.
			New(path.Join(outputRoot, relativePath)).
			Funcs(Funcs).
			Parse(path.Join(outputRoot, replacedPath))
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, cfg)
		if err != nil {
			return err
		}
		outputPath := buf.String()

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"path": p,
			}).Error("Unable to access file")
		} else if info.IsDir() {
			return handleDirectory(templatePath, outputPath)
		} else {
			return handleFile(templatePath, outputPath, cfg)
		}
		return nil
	}
}

func handleDirectory(templatePath, outputPath string) error {
	log.WithFields(log.Fields{
		"path": outputPath,
	}).Debug("Creating directory")
	return os.MkdirAll(outputPath, 0777)
}

func handleFile(templatePath, outputPath string, cfg *Config) error {
	log.WithFields(log.Fields{
		"path": outputPath,
	}).Debug("Expanding template")

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Make every .sh file executable
	if path.Ext(outputPath) == ".sh" {
		stat, err := f.Stat()
		if err != nil {
			return err
		}

		err = f.Chmod(stat.Mode() | 0700)
		if err != nil {
			return err
		}
	}

	tmpl, err := getTemplate(templatePath)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, cfg)
}

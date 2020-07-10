// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	pschema "github.com/pulumi/pulumi/pkg/v2/codegen/schema"
	templatesgen "github.com/pulumi/pulumi/pkg/v2/codegen/templates"
	"github.com/pulumi/pulumi/sdk/v2/go/common/tools"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/contract"
)

func main() {
	// Grab the args.
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "error: usage: %s <out-dir> <provider-schema-file>\n", os.Args[0])
		os.Exit(1)
	}

	defer glog.Flush()

	outDir, schemaFile := args[0], args[1]

	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		glog.Infof("error reading schema file from path: %v", err)
		os.Exit(1)
	}

	mainSpec := &pschema.PackageSpec{}
	if err := json.Unmarshal(schema, mainSpec); err != nil {
		glog.Infof("error unmarshalling schema into a PackageSpec: %v", err)
		os.Exit(1)
	}

	if err := generateDocsFromSchema(outDir, mainSpec); err != nil {
		glog.Infof("error generating docs from schema: %v", err)
		os.Exit(1)
	}
}

func generateDocsFromSchema(outDir string, spec *pschema.PackageSpec) error {
	pulPkg, err := pschema.ImportSpec(*spec, nil)
	if err != nil {
		return errors.Wrapf(err, "error importing package spec: %v", err)
	}

	files, err := templatesgen.GeneratePackage("Pulumi Templates Generator", pulPkg)
	if err != nil {
		return errors.Wrap(err, "generating Pulumi package")
	}

	for f, contents := range files {
		if err := emitFile(outDir, f, contents); err != nil {
			return errors.Wrapf(err, "emitting file %v", f)
		}
	}
	return nil
}

func emitFile(outDir, relPath string, contents []byte) error {
	p := path.Join(outDir, relPath)
	if err := tools.EnsureDir(path.Dir(p)); err != nil {
		return errors.Wrap(err, "creating directory")
	}

	f, err := os.Create(p)
	if err != nil {
		return errors.Wrap(err, "creating file")
	}
	defer contract.IgnoreClose(f)

	_, err = f.Write(contents)
	return err
}

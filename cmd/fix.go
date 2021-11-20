/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type ImportArrays struct {
	importStrings         []string
	dartStrings           []string
	flutterStrings        []string
	thirdPartyImports     []string
	routesImports         []string
	utilImports           []string
	sharedImports         []string
	blocImports           []string
	modulesImports        []string
	otherImports          []string
	commentsAndDirectives []string
	libraryString         []string
}

// fixCmd represents the fix command
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Fix dart directory use -i option ",
	Long:  `Helps in ordering imports`,
	Run: func(cmd *cobra.Command, args []string) {
		fstatus, _ := cmd.Flags().GetBool("import")

		if fstatus { // if status is true, call addFloat
			fixImportOrder(args)
		} else {
			fmt.Println("fix called without i")
		}
	},
}

func init() {
	rootCmd.AddCommand(fixCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fixCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fixCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	fixCmd.Flags().BoolP("import", "i", false, "Fix import order")
}

func fixImportOrder(args []string) {
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".dart") {
				fmt.Println(path, info.Size())

				readAndFixImports(path)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func readAndFixImports(path string) {
	arrs := ImportArrays{}

	rex := regexp.MustCompile("//(.)?@(.)*")

	re := regexp.MustCompile(`(//)*(.)?import '(.)*:(.)*'(.)*;(\s)*`)

	libex := regexp.MustCompile("library(.)*;")

	bs, err := ioutil.ReadFile(path)

	if err != nil {

		log.Fatal(err)
	}

	text := string(bs)

	arrs.importStrings = re.FindAllString(text, -1)

	arrs.commentsAndDirectives = rex.FindAllString(text, -1)

	arrs.libraryString = libex.FindAllString(text, -1)

	arrs.sortAllImports()

	var sb strings.Builder

	buildImports(arrs.commentsAndDirectives, &sb)
	buildImports(arrs.libraryString, &sb)
	buildImports(arrs.dartStrings, &sb)
	buildImports(arrs.flutterStrings, &sb)
	buildImports(arrs.thirdPartyImports, &sb)
	buildImports(arrs.otherImports, &sb)
	buildImports(arrs.routesImports, &sb)
	buildImports(arrs.utilImports, &sb)
	buildImports(arrs.sharedImports, &sb)
	buildImports(arrs.blocImports, &sb)
	buildImports(arrs.modulesImports, &sb)

	text = re.ReplaceAllString(text, "")

	text = rex.ReplaceAllString(text, "")

	text = libex.ReplaceAllString(text, "")

	text = strings.TrimSpace(text)

	sb.WriteString(text)

	ioutil.WriteFile(path, []byte(sb.String()), 0)

}

func buildImports(arr []string, sb *strings.Builder) {
	for _, str := range arr {
		sb.WriteString(str)
		sb.WriteString("\n")
	}
	if len(arr) >= 1 {
		sb.WriteString("\n")
	}
}

func (arr *ImportArrays) sortAllImports() {

	for _, str := range arr.importStrings {

		str = strings.TrimSpace(str)

		if isDartImport(str) {
			arr.dartStrings = append(arr.dartStrings, str)
			sort.Slice(arr.dartStrings, func(p, q int) bool {
				return arr.dartStrings[p] < arr.dartStrings[q]
			})
			continue
		}

		if isFlutterImport(str) {
			arr.flutterStrings = append(arr.flutterStrings, str)
			sort.Slice(arr.flutterStrings, func(p, q int) bool {
				return arr.flutterStrings[p] < arr.flutterStrings[q]
			})
			continue
		}

		if isThirdPartyImport(str) {
			arr.thirdPartyImports = append(arr.thirdPartyImports, str)
			sort.Slice(arr.thirdPartyImports, func(p, q int) bool {
				return arr.thirdPartyImports[p] < arr.thirdPartyImports[q]
			})
			continue
		}
		if matchImport(str, "package:yc_app/routes") {
			arr.routesImports = append(arr.routesImports, str)
			sort.Slice(arr.routesImports, func(p, q int) bool {
				return arr.routesImports[p] < arr.routesImports[q]
			})
			continue
		}
		if matchImport(str, "package:yc_app/utils/") {
			arr.utilImports = append(arr.utilImports, str)
			sort.Slice(arr.utilImports, func(p, q int) bool {
				return arr.utilImports[p] < arr.utilImports[q]
			})
			continue
		}
		if matchImport(str, "package:yc_app/shared/") {
			arr.sharedImports = append(arr.sharedImports, str)
			sort.Slice(arr.sharedImports, func(p, q int) bool {
				return arr.sharedImports[p] < arr.sharedImports[q]
			})
			continue
		}
		if matchImport(str, "package:yc_app/blocs/") {
			arr.blocImports = append(arr.blocImports, str)
			sort.Slice(arr.blocImports, func(p, q int) bool {
				return arr.blocImports[p] < arr.blocImports[q]
			})
			continue
		}
		if matchImport(str, "package:yc_app/modules/") {
			arr.modulesImports = append(arr.modulesImports, str)
			sort.Slice(arr.modulesImports, func(p, q int) bool {
				return arr.modulesImports[p] < arr.modulesImports[q]
			})
			continue
		}

		arr.otherImports = append(arr.otherImports, str)
		sort.Slice(arr.otherImports, func(p, q int) bool {
			return arr.otherImports[p] < arr.otherImports[q]
		})

	}
}

func isDartImport(str string) bool {
	re := regexp.MustCompile("^import(.)*'dart:(.)*;(/s)*")
	strs := re.FindString(str)
	if strs != "" {
		return true
	}
	return false
}

func isCommentImport(str string) bool {
	re := regexp.MustCompile("//(.)*import(/s)*;(/s)*")
	strs := re.FindString(str)
	if strs != "" {
		return true
	}
	return false
}

func isThirdPartyImport(str string) bool {
	if !(strings.Contains(str, "package:yc_app/") || strings.Contains(str, "package:flutter/")) && strings.Contains(str, "package:") {
		return true
	}
	return false
}
func isFlutterImport(str string) bool {
	re, _ := regexp.Compile("^(import 'package:flutter/)")
	strs := re.FindString(str)
	if strs != "" {
		return true
	}
	return false
}

func matchImport(str string, condition string) bool {
	if strings.Contains(str, condition) {
		return true
	}
	return false
}

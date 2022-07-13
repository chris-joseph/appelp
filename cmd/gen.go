package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func getDartTypeList() []string {
	return []string{"int", "double", "String", "bool", "dynamic"}
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate dart classes from .json files in the directory",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		generator(args)
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
	},
}

func generator(args []string) {
	fmt.Println("hey gen user")
	var total int
	start := time.Now()
	go startCrawl(concurrency)
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".json") {
				fmt.Println(path, info.Size())

				jobs <- path
				total++

			}
			return nil
		})
	wg.Wait()
	close(jobs)
	fmt.Printf("\nThreads: %d\n", concurrency)
	fmt.Printf("Total: %d\n", total)
	fmt.Printf("Time: %s\n", time.Since(start).String())
	if err != nil {
		log.Println(err)
	}
}
func startCrawl(threads int) {
	for i := 0; i < threads; i++ {
		go func() {
			for {
				path, ok := <-jobs
				if !ok {
					return
				}
				wg.Add(1)
				fmt.Println("WORKING ON --- ", path)
				readJson(path)
				fmt.Println("FINISHED ---- ", path)
				wg.Done()
			}
		}()
	}
}
func readJson(path string) {
	var sb strings.Builder
	bs, err := ioutil.ReadFile(path)

	if err != nil {

		log.Fatal(err)
	}

	text := string(bs)
	valid := json.Valid([]byte(text))
	if !valid {
		return
	}

	dartPath := strings.Split(path, ".")
	dartPath[len(dartPath)-1] = ""
	caser := cases.Title(language.Und)
	jsonParse(text, caser.String(strings.ToLower(strings.Join(dartPath, ""))), &sb)
	dartPath[len(dartPath)-1] = ".dart"
	ioutil.WriteFile(strings.Join(dartPath, ""), []byte(sb.String()), 0)
}
func jsonParse(str string, name string, sb *strings.Builder) {
	var output map[string]any
	var params = make(map[string]string)

	json.Unmarshal([]byte(str), &output)

	fmt.Println(output)
	for key, val := range output {
		params[key] = getType(val, key, sb)
	}
	classBuilder(sb, name, params)
}
func getType(val any, key string, sb *strings.Builder) string {
	k := reflect.ValueOf(val).Kind().String()

	switch k {

	case "bool":
		return "bool"
	case "float32", "float64":
		res := strconv.FormatFloat(val.(float64), 'f', 6, 64)
		v, _ := strconv.ParseFloat(res, 64)
		if val == float64(int(v)) {
			return "int"
		}
		return "double"
	case "map":
		empData, err := json.Marshal(val)
		if err != nil {
			fmt.Println(err.Error())
			return "Map<dynamic,dynamic>"

		}
		caser := cases.Title(language.Und)
		jsonStr := string(empData)
		jsonParse(jsonStr, caser.String((strings.ToLower(key))), sb)
		return caser.String(strings.ToLower(key))

	case "slice":
		j := val.([]any)
		//TODO add logic to sepereate double and int by iterating all values
		t := getType(j[0], key, sb)

		return "List<" + t + ">"
	case "string":

		_, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", val))
		if err == nil {
			return "DateTime"
		}

		return "String"
	default:
		return "dynamic"
	}
}
func classBuilder(sb *strings.Builder, name string, params map[string]string) {
	sb.WriteString("\n")
	sb.WriteString("\n")
	sb.WriteString(`class ` + name + ` {`)
	sb.WriteString("\n")
	for k, v := range params {
		sb.WriteString("  final " + v + "? " + k + ";")
		sb.WriteString("\n")
	}
	sb.WriteString("  " + makeConstructor(name, params))
	sb.WriteString("\n")
	sb.WriteString("  " + makeFromJson(name, params))
	sb.WriteString("\n")
	sb.WriteString("  " + makeToJson(name, params))
	sb.WriteString("\n")
	sb.WriteString("  " + makeCopyWith(name, params))
	sb.WriteString("\n")
	sb.WriteString("}")
	sb.WriteString("\n")
	print(sb.String())

}

func makeConstructor(name string, params map[string]string) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("const " + name + "({")
	for k := range params {
		sb.WriteString(" this." + k + ",")

	}
	sb.WriteString("});")
	return sb.String()

}

func makeFromJson(name string, params map[string]string) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("factory " + name + ".fromJson(Map<String, dynamic> json)=>" + "\n")
	sb.WriteString(name + "(")
	sb.WriteString("\n")
	for k, v := range params {
		if strings.Contains(v, "List") {
			sb.WriteString(makeListFromJson(k, v))
			continue
		}
		if strings.Contains(v, "DateTime") {
			sb.WriteString(" " + k + ":" + "DateTime.parse(json[\"" + k + "\"])" + ",")
			continue
		}
		if !slices.Contains(getDartTypeList(), v) {
			sb.WriteString(" " + k + ":" + v + ".fromJson(json[\"" + k + "\"])" + ",")
			sb.WriteString("\n")
			continue
		}
		sb.WriteString(" " + k + ":" + "json[\"" + k + "\"]" + ",")
		sb.WriteString("\n")
	}

	sb.WriteString(");")
	return sb.String()

}

func makeListFromJson(k string, v string) string {
	j := strings.Split(v, "<")
	j = j[1:]
	j = strings.Split(strings.Join(j, ""), ">")
	typ := strings.Join(j, "")
	if slices.Contains(getDartTypeList(), typ) {
		return " " + k + ":" + "json[\"" + k + "\"]" + "," + "\n"
	}

	return " " + k + ":" + v + ".from(json[\"" + k + "\"].map((x) => " + typ + ".fromJson(x)))" + "," + "\n"

}

func makeListToJson(k string, v string) string {
	j := strings.Split(v, "<")
	j = j[1:]
	j = strings.Split(strings.Join(j, ""), ">")
	typ := strings.Join(j, "")
	if slices.Contains(getDartTypeList(), typ) {
		return " \"" + k + "\":" + k + "," + "\n"
	}

	return " \"" + k + "\":" + k + "==null?null:List<dynamic>" + ".from(" + k + "!.map((x) => " + "x.toJson()))" + "," + "\n"

}

func makeCopyWith(name string, params map[string]string) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(name + " copyWith({" + "\n")
	for k, v := range params {

		sb.WriteString(" " + v + "?" + k + ",")
		sb.WriteString("\n")

	}
	sb.WriteString("})=>" + name + "(\n")
	for k := range params {

		sb.WriteString(" " + k + ":" + k + "?? this." + k + ",")
		sb.WriteString("\n")

	}
	sb.WriteString(");\n")
	return sb.String()
}

func makeToJson(name string, params map[string]string) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("Map<String, dynamic> toJson(){" + "\n")
	sb.WriteString("var map= {" + "\n")

	for k, v := range params {
		if strings.Contains(v, "List") {
			sb.WriteString(makeListToJson(k, v))
			continue
		}
		if strings.Contains(v, "DateTime") {
			sb.WriteString(" \"" + k + "\":" + k + "?.toIso8601String()" + ",")
			continue
		}
		if !slices.Contains(getDartTypeList(), v) {
			sb.WriteString(" \"" + k + "\":" + k + "?.toJson()" + ",")
			sb.WriteString("\n")
			continue
		}
		sb.WriteString(" \"" + k + "\":" + k + ",")
		sb.WriteString("\n")

	}
	sb.WriteString("};\n")
	sb.WriteString("map.removeWhere((key, value) => value == null);\n")
	sb.WriteString("return map;\n")
	sb.WriteString("}\n")
	return sb.String()
}

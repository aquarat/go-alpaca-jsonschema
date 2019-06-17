package alpacajsonschema

import (
	"log"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

type JSONSchema map[string]interface{}

type FormMeta struct {
	Title       string
	Description string
}

func RenderSchema(t interface{}, fmd FormMeta) JSONSchema {
	v := reflect.ValueOf(t)
	vt := reflect.TypeOf(t)

	schemaObject := make(map[string]interface{})
	schema := make(map[string]interface{})
	schemaProps := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		tag := vt.Field(i).Tag.Get("jsonschema")

		log.Println("source tag", tag, " fieldname ", vt.Field(i).Name)

		schema["title"] = fmd.Title
		schema["description"] = fmd.Description

		if len(tag) > 0 {
			switch vt.Field(i).Name {
			default:
				prop := make(map[string]interface{})

				fieldName := vt.Field(i).Name
				tit, ok := vt.Field(i).Tag.Lookup("jsonschematitle")
				if ok {
					fieldName = tit
				}
				if strings.Contains(tag, "enum") {
					prop["enum"] = getFieldOptions(tag, v, vt)

					switch vt.Field(i).Type.Kind() {
					case reflect.Int:
						prop["type"] = "number"
					case reflect.Float64:
						prop["type"] = "number"
					default:
						prop["type"] = "string"
					}
				} else {
					switch vt.Field(i).Type.Kind() {
					case reflect.Int:
						prop["type"] = "number"
					case reflect.Float64:
						prop["type"] = "number"
					case reflect.Bool:
						prop["type"] = "boolean"
					default:
						prop["type"] = "string"
					}
				}

				if strings.Contains(tag, "required") {
					prop["required"] = true
				} else {
					prop["required"] = false
				}

				fieldName = spaceFieldName(fieldName)
				prop["title"] = fieldName

				schemaProps[strings.ToLower(vt.Field(i).Name)] = prop
			}
		}
	}

	schema["type"] = "object"
	schema["properties"] = schemaProps
	schemaObject["schema"] = schema

	return schemaObject
}

func spaceFieldName(a string) (fieldName string) {
	words := Split(a)
	for i, j := range words {
		if i > 0 {
			fieldName = fieldName + " " + j
		} else {
			fieldName = j
		}
	}

	return
}

func getTagValue(tags, tag string) (contents string) {
	log.Println("tags ", tags)
	for _, j := range strings.Split(tag, ",") {
		log.Println(j)

		if strings.Contains(j, tag) {
			contents = strings.Trim(j, "\"")

		}
	}

	return
}

func getFieldOptions(tag string, val reflect.Value, typ reflect.Type) (OptionsField []string) {
	tags := strings.Split(tag, ",")
	optionsFieldIdent := "OptionsField"

	optionsFieldName := ""

	for _, j := range tags {
		if strings.Contains(j, optionsFieldIdent) {
			if len(j) > 12 {
				optionsFieldName = j[len(optionsFieldIdent)+1:]
				break
			}
		}
	}

	for i := 0; i < val.NumField(); i++ {
		if typ.Field(i).Name == optionsFieldName {
			OptionsField = strings.Split(val.Field(i).String(), ",")
		}
	}

	return
}

//https://github.com/fatih/camelcase/blob/master/camelcase.go
// Split splits the camelcase word and returns a list of words. It also
// supports digits. Both lower camel case and upper camel case are supported.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
//
// Examples
//
//   "" =>                     [""]
//   "lowercase" =>            ["lowercase"]
//   "Class" =>                ["Class"]
//   "MyClass" =>              ["My", "Class"]
//   "MyC" =>                  ["My", "C"]
//   "HTML" =>                 ["HTML"]
//   "PDFLoader" =>            ["PDF", "Loader"]
//   "AString" =>              ["A", "String"]
//   "SimpleXMLParser" =>      ["Simple", "XML", "Parser"]
//   "vimRPCPlugin" =>         ["vim", "RPC", "Plugin"]
//   "GL11Version" =>          ["GL", "11", "Version"]
//   "99Bottles" =>            ["99", "Bottles"]
//   "May5" =>                 ["May", "5"]
//   "BFG9000" =>              ["BFG", "9000"]
//   "BöseÜberraschung" =>     ["Böse", "Überraschung"]
//   "Two  spaces" =>          ["Two", "  ", "spaces"]
//   "BadUTF8\xe2\xe2\xa1" =>  ["BadUTF8\xe2\xe2\xa1"]
//
// Splitting rules
//
//  1) If string is not valid UTF-8, return it without splitting as
//     single item array.
//  2) Assign all unicode characters into one of 4 sets: lower case
//     letters, upper case letters, numbers, and all other characters.
//  3) Iterate through characters of string, introducing splits
//     between adjacent characters that belong to different sets.
//  4) Iterate through array of split strings, and if a given string
//     is upper case:
//       if subsequent string is lower case:
//         move last character of upper case string to beginning of
//         lower case string
func Split(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return
}

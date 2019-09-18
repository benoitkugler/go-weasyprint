package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
)

func loadJson(filename string) ([]string, []string) {
	b, err := ioutil.ReadFile(filepath.Join("css-parsing-tests", filename))
	if err != nil {
		log.Fatal(err)
	}
	var l []interface{}
	if err = json.Unmarshal(b, &l); err != nil {
		log.Fatal(err)
	}
	if len(l)%2 != 0 {
		log.Fatal("number of tests in list should be even !")
	}
	inputs, resJsons := make([]string, len(l)/2), make([]string, len(l)/2)
	for i := 0; i < len(l); i += 2 {
		inputs[i/2] = l[i].(string)
		res, err := json.Marshal(l[i+1])
		if err != nil {
			log.Fatal(err)
		}
		resJsons[i/2] = string(res)
	}
	return inputs, resJsons
}

func runTest(t *testing.T, css, resJson []string, fn func(input string) []Token) {
	for i, input := range css {
		resToTest := fn(input)
		res, err := marshalJSON(resToTest)
		if err != nil {
			t.Fatal(err)
		}
		if res != resJson[i] {
			t.Errorf(fmt.Sprintf("input (%d) %s faile d \n expected %s \n got  \n %s \n", i, input, resJson[i], res))
		}
	}
}

func TestComponentValueList(t *testing.T) {
	inputs, resJson := loadJson("component_value_list.json")
	// l := ParseComponentValueList(inputs[6], true)
	// fmt.Println(toJson(l))
	// fmt.Println(resJson[6], "uu")
	// fmt.Println(marshalJSON(l))
	runTest(t, inputs, resJson, func(s string) []Token {
		return ParseComponentValueList(s, true)
	})
}

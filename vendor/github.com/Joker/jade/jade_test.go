package jade

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJadeExamples(t *testing.T) {

	files, _ := ioutil.ReadDir("./testdata")
	var name, fext string

	for _, file := range files {
		name = file.Name()
		fext = filepath.Ext(name)

		if fext != ".jade" && fext != ".pug" {
			continue
		}

		fmt.Println("_________" + name)

		dat, err := ioutil.ReadFile("testdata/" + name)
		if err != nil {
			fmt.Printf("--- FAIL: ReadFile error: %v\n\n", err)
			t.Fail()
			continue
		}

		tpl, err := Parse(name, string(dat))
		if err != nil {
			fmt.Printf("--- FAIL: Parse error: %v\n\n", err)
			t.Fail()
			continue
		}
		tmpl := bufio.NewScanner(strings.NewReader(tpl))
		tmpl.Split(bufio.ScanLines)

		inFile, err := os.Open("testdata/" + strings.TrimSuffix(name, fext) + ".html")
		if err != nil {
			fmt.Println("```", tpl, "\n\n```")
			continue
		}
		html := bufio.NewScanner(inFile)
		html.Split(bufio.ScanLines)

		nilerr, line := 0, 0

		for tmpl.Scan() {
			html.Scan()

			a := tmpl.Text()
			b := html.Text()
			line += 1

			if strings.Compare(a, b) != 0 && nilerr < 4 {
				fmt.Printf("%s\n%s\n%d^___________________________\n", a, b, line)
				nilerr += 1
				t.Fail()
			}
		}
		inFile.Close()

		if nilerr == 0 {
			fmt.Println("    PASS\n")
		} else {
			fmt.Println("--- FAIL\n")
		}
	}
	// ioutil.WriteFile("testdata/"+name+".html", []byte(tpl), 0644)
}

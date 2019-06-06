package main

import (
	"fmt"
	"io"
	"os"
	//"path/filepath"
	"strings"
	"sort"
	"io/ioutil"
)

func printTree(output io.Writer, path string, isPrintFiles bool, deep string) (err error) {
	err = nil
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	} 

	buf := make([]os.FileInfo, 0, len(files))
	for _, val := range files {
		if (val.IsDir() && !isPrintFiles) {
			buf = append(buf, val)
		} else if (isPrintFiles) {
			buf = append(buf, val)
		}
	}

	sort.Slice(buf, func(i, j int) bool {
		if (strings.Compare(buf[i].Name(), buf[j].Name()) >= 0) {
			return false
		}
		return true
	})
	sep := "├───"
	deepSep := "│\t"
	for key, val := range buf {
		if key + 1 == len(buf) {
			sep = "└───"
			deepSep = "\t"
		}
		if val.IsDir() {
			_, err = fmt.Fprintf(output, "%s%s%s\n", deep, sep, val.Name())
			printTree(output, path + string(os.PathSeparator) + val.Name(), isPrintFiles, deep + deepSep)
		} else {
			if (val.Size() > 0) {
				_, err = fmt.Fprintf(output, "%s%s%s (%db)\n", deep, sep, val.Name(), val.Size())
			} else {
				_, err = fmt.Fprintf(output, "%s%s%s (empty)\n", deep, sep, val.Name())
			}
		}
	}
	return err
}

func dirTree (output io.Writer, path string, isPrintFiles bool) (err error) {
	return printTree(output, path, isPrintFiles, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

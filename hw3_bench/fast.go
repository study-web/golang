package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

// вам надо написать более быструю оптимальную этой функции
// func FastSearch(out io.Writer) {
// 	SlowSearch(out)
// }

var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

type User struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"company"`
	Country  string   `json:"country"`
	Email    string   `json:"email"`
	Job      string   `json:"job"`
	Name     string   `json:"name"`
	Phone    string   `json:"phone"`
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(file)

	r := strings.NewReplacer("@", " [at] ")
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	userT := User{}

	fmt.Fprintln(out, "found users:")
	i := 0
	var line = make([]byte, 1024, 1024)
LOOP:
	for {
		i++
		j := 0
		for {
			b, err := reader.ReadByte()
			if b == '\n' {
				break
			}
			if err != nil {
				if err == io.EOF {
					break LOOP
				} else {
					panic(err)
				}
			}
			line[j] = b
			j++
		}
		if !bytes.Contains(line[:j], []byte("Android")) && !bytes.Contains(line[:j], []byte("MSIE")) {
			continue
		}
		err = userT.UnmarshalJSON(line[:j])
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range userT.Browsers {
			if ok := strings.Contains(browser, "Android"); ok {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browser := range userT.Browsers {
			if ok := strings.Contains(browser, "MSIE"); ok {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}
		email := r.Replace(userT.Email)
		fmt.Fprintln(out, fmt.Sprintf("[%d] %s <%s>", i-1, userT.Name, email))
	}
	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}

func easyjson3486653aDecodeCUsersAlexGoSrc(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "company":
			out.Company = string(in.String())
		case "country":
			out.Country = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "job":
			out.Job = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "phone":
			out.Phone = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3486653aDecodeCUsersAlexGoSrc(&r, v)
	return r.Error()
}

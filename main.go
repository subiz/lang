package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/leonelquinteros/gotext"
)

func main() {
	if err := buildAllPoInPath("./widget/"); err != nil {
		panic(err)
	}
	if err := buildAllPoInPath("./dashboard/"); err != nil {
		panic(err)
	}
}

func buildAllPoInPath(path string) error {
	// read all po
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	en, err := ioutil.ReadFile(path + "en-US.po")
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if f.Name() == "en-US.po" {
			continue
		}

		if !strings.HasSuffix(f.Name(), ".po") {
			continue
		}

		content, err := ioutil.ReadFile(path + f.Name())
		if err != nil {
			return err
		}

		newcontent := Merge(en, content)
		if err := ioutil.WriteFile(path+f.Name(), []byte(newcontent), 0644); err != nil {
			return err
		}

		fmt.Println("written to", path+f.Name())
		name := (path + f.Name())[:len(path+f.Name())-3]
		// generate json
		js := Jsonify(ParsePO(newcontent))

		if err := ioutil.WriteFile(name + ".json", []byte(js), 0644); err != nil {
			return err
		}
		fmt.Println("written to", name + ".json")

		if err := ioutil.WriteFile(name + ".js", []byte("export default " + js), 0644); err != nil {
			return err
		}
		fmt.Println("written to", name + ".js")
	}

	name := path + "en-US"
	js := Jsonify(ParsePO(en))
	if err := ioutil.WriteFile(name + ".json", []byte(js), 0644); err != nil {
		return err
	}
	fmt.Println("written to", name + ".json")

	if err := ioutil.WriteFile(name + ".js", []byte("export default " + js), 0644); err != nil {
		return err
	}
	fmt.Println("written to", name + ".js")
	fmt.Println("done.")
	return nil
}

// ------------------------------------------------
type PoElement struct {
	Context, Id, Str string
}

// ParsePO converts .po to array of PoElement for future manipulations
func ParsePO(pocontent []byte) []PoElement {
	po := &gotext.Po{}
	po.Parse(pocontent)

	// here, we want getting all .po Contexts
	// but the library do not provide any way to directly get it from gotext.Po,
	// we have to marshal then unmarshal the Po to gotext.TranslatorEncoding to
	// get the Contexts.
	data, _ := po.MarshalBinary()
	obj := new(gotext.TranslatorEncoding)
	decoder := gob.NewDecoder(bytes.NewBuffer(data))
	if err := decoder.Decode(obj); err != nil {
		return nil
	}

	// hold the output
	poes := make([]PoElement, 0)
	for c, ctx := range obj.Contexts {
		for id, st := range ctx {
			if id == "" {
				continue
			}

			str := ""
			for i, r := range st.Trs {
				if i != 0 {
					str += " | "
				}
				str += r
			}

			poes = append(poes, PoElement{Context: c, Id: id, Str: str})
		}
	}
	return poes
}

// espaceSlash converts slashes to PO representations
func escapeSlash(str string) string {
	return strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(str, "\n", "\\n", -1),
				"\r", "\\r", -1),
			"\t", "\\t", -1),
		"\"", "\\\"", -1)
}

// poify converts arrays of PoElement to .po format
func Poify(eles []PoElement) []byte {
	sort(eles)

	var buffer bytes.Buffer
	for _, e := range eles {
		buffer.WriteString(fmt.Sprintf(`msgctxt "%s"
msgid "%s"
msgstr "%s"

`, escapeSlash(e.Context), escapeSlash(e.Id), escapeSlash(e.Str)))
	}
	return buffer.Bytes()
}

func sort(eles []PoElement) {
	for i := 0; i < len(eles); i++ {
		for j := i + 1; j < len(eles); j++ {
			if eles[j].Context < eles[i].Context {
				eles[i], eles[j] = eles[j], eles[i]
			} else if eles[i].Context == eles[j].Context {
				if eles[j].Id < eles[i].Id {
					eles[i], eles[j] = eles[j], eles[i]
				}
			}
		}
	}
}

func toJsonKey(key string) string {
	key = strings.TrimSpace(key)
	if key[0] == '.' {
		return key[1:]
	}
	return key
}

// jsonifyPo converts array of PoElement to JSON representation
func Jsonify(eles []PoElement) string {
	if len(eles) == 0 {
		return ""
	}
	sort(eles)
	var buffer bytes.Buffer
	buffer.WriteString("{\n")
	for i, e := range eles {
		jsonkey := strings.TrimSpace(e.Context)
		if e.Context[0] == '.' {
			jsonkey = e.Context[1:]
		}

		buffer.WriteString(fmt.Sprintf(`	"%s": "%s"`, jsonkey, escapeSlash(e.Str)))
		if i == len(eles)-1 {
			buffer.WriteString("\n}")
		} else {
			buffer.WriteString(",\n")
		}
	}
	return buffer.String()
}

func JsonToPo(en, la []byte) ([]byte, error) {
	var enmap map[string]string
	var lamap map[string]string

	if err := json.Unmarshal(en, &enmap); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(la, &lamap); err != nil {
		return nil, err
	}

	poes := make([]PoElement, 0)
	for k, v := range enmap {
		e := PoElement{
			Context: "." + k,
			Id:      v,
			Str:     lamap[k],
		}
		poes = append(poes, e)
	}
	return Poify(poes), nil
}

func Merge(enpo, newpo []byte) []byte {
	eneles := ParsePO(enpo)
	neweles := ParsePO(newpo)
	eles := make([]PoElement, 0)
	for _, ee := range eneles {
		newe := PoElement{Id: ee.Id, Context: ee.Context, Str: ee.Str}
		for _, ne := range neweles {
			if ee.Context == ne.Context && ee.Id == ne.Id {
				newe.Str = ne.Str
			}
		}

		eles = append(eles, newe)
	}
	return Poify(eles)
}

package templates

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"text/template"
	"time"

	"DreamsMoney/feelgoodfoodsnv.com/ordering/config"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

var serverConfiguration config.Config

func RunTemplateOutputInterval(config config.Config, syncer chan struct{}) {
	serverConfiguration = config
	atStart()
	go func() {
		for {
			everyTwoMinutes()
			time.Sleep(time.Minute * 2)
			syncer <- struct{}{}
		}
	}()
}

func mapArgs(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid arg call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("arg keys must be strings")
		}
		dict[key] = values[i+1]
	}

	dict["environment"] = serverConfiguration.Environment

	return dict, nil
}

func minification(filenames ...string) (*template.Template, error) {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/js", js.Minify)

	var tmpl *template.Template
	for _, filename := range filenames {
		name := filepath.Base(filename)
		if tmpl == nil {
			tmpl = template.New(name).Funcs(template.FuncMap{"args": mapArgs})
		} else {
			tmpl = tmpl.New(name).Funcs(template.FuncMap{"args": mapArgs})
		}

		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		mb, err := m.Bytes("text/html", b)
		if err != nil {
			return nil, err
		}

		tmpl.Parse(string(mb))
	}
	return tmpl, nil
}

func ParseFiles(files ...string) *template.Template {
	return template.Must(minification(files...))
}

// func ParseFiles(files ...string) *template.Template {
// 	return template.Must(template.New("").Funcs(template.FuncMap{"args": mapArgs}).ParseFiles(files...))
// }

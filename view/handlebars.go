// Copyright 2017 Gerasimos Maropoulos, ΓΜ. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aymerick/raymond"
)

// HandlebarsEngine contains the handlebars view engine structure.
type HandlebarsEngine struct {
	// files configuration
	directory string
	extension string
	assetFn   func(name string) ([]byte, error) // for embedded, in combination with directory & extension
	namesFn   func() []string                   // for embedded, in combination with directory & extension
	reload    bool                              // if true, each time the ExecuteWriter is called the templates will be reloaded.
	// parser configuration
	layout        string
	rmu           sync.RWMutex // locks for helpers
	helpers       map[string]interface{}
	mu            sync.Mutex // locks for template files load
	templateCache map[string]*raymond.Template
}

// Handlebars creates and returns a new handlebars view engine.
func Handlebars(directory, extension string) *HandlebarsEngine {
	s := &HandlebarsEngine{
		directory:     directory,
		extension:     extension,
		templateCache: make(map[string]*raymond.Template, 0),
		helpers:       make(map[string]interface{}, 0),
	}

	// register the render helper here
	raymond.RegisterHelper("render", func(partial string, binding interface{}) raymond.SafeString {
		contents, err := s.executeTemplateBuf(partial, binding)
		if err != nil {
			return raymond.SafeString("template with name: " + partial + " couldn't not be found.")
		}
		return raymond.SafeString(contents)
	})

	return s
}

// Ext returns the file extension which this view engine is responsible to render.
func (s *HandlebarsEngine) Ext() string {
	return s.extension
}

// Binary optionally, use it when template files are distributed
// inside the app executable (.go generated files).
//
// The assetFn and namesFn can come from the go-bindata library.
func (s *HandlebarsEngine) Binary(assetFn func(name string) ([]byte, error), namesFn func() []string) *HandlebarsEngine {
	s.assetFn, s.namesFn = assetFn, namesFn
	return s
}

// Reload if set to true the templates are reloading on each render,
// use it when you're in development and you're boring of restarting
// the whole app when you edit a template file.
func (s *HandlebarsEngine) Reload(developmentMode bool) *HandlebarsEngine {
	s.reload = developmentMode
	return s
}

// Layout sets the layout template file which should use
// the {{ yield }} func to yield the main template file
// and optionally {{partial/partial_r/render}} to render
// other template files like headers and footers.
func (s *HandlebarsEngine) Layout(layoutFile string) *HandlebarsEngine {
	s.layout = layoutFile
	return s
}

// AddFunc adds the function to the template's function map.
// It is legal to overwrite elements of the default actions:
// - url func(routeName string, args ...string) string
// - urlpath func(routeName string, args ...string) string
// - render func(fullPartialName string) (raymond.HTML, error).
func (s *HandlebarsEngine) AddFunc(funcName string, funcBody interface{}) {
	s.rmu.Lock()
	s.helpers[funcName] = funcBody
	s.rmu.Unlock()
}

// Load parses the templates to the engine.
// It's alos responsible to add the necessary global functions.
//
// Returns an error if something bad happens, user is responsible to catch it.
func (s *HandlebarsEngine) Load() error {
	if s.assetFn != nil && s.namesFn != nil {
		// embedded
		return s.loadAssets()
	}

	// load from directory, make the dir absolute here too.
	dir, err := filepath.Abs(s.directory)
	if err != nil {
		return err
	}
	// change the directory field configuration, load happens after directory has been set, so we will not have any problems here.
	s.directory = dir
	return s.loadDirectory()
}

// loadDirectory builds the handlebars templates from directory.
func (s *HandlebarsEngine) loadDirectory() error {

	// register the global helpers on the first load
	if len(s.templateCache) == 0 && s.helpers != nil {
		raymond.RegisterHelpers(s.helpers)
	}

	dir, extension := s.directory, s.extension

	// the render works like {{ render "myfile.html" theContext.PartialContext}}
	// instead of the html/template engine which works like {{ render "myfile.html"}} and accepts the parent binding, with handlebars we can't do that because of lack of runtime helpers (dublicate error)
	s.mu.Lock()
	defer s.mu.Unlock()
	var templateErr error
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := filepath.Ext(rel)
		if ext == extension {
			buf, err := ioutil.ReadFile(path)
			contents := string(buf)

			if err != nil {
				templateErr = err
				return err
			}

			name := filepath.ToSlash(rel)

			tmpl, err := raymond.Parse(contents)
			if err != nil {
				templateErr = err
				return err
			}
			s.templateCache[name] = tmpl
		}
		return nil
	})

	return templateErr
}

// loadAssets loads the templates by binary, embedded.
func (s *HandlebarsEngine) loadAssets() error {
	// register the global helpers
	if len(s.templateCache) == 0 && s.helpers != nil {
		raymond.RegisterHelpers(s.helpers)
	}

	virtualDirectory, virtualExtension := s.directory, s.extension
	assetFn, namesFn := s.assetFn, s.namesFn

	if len(virtualDirectory) > 0 {
		if virtualDirectory[0] == '.' { // first check for .wrong
			virtualDirectory = virtualDirectory[1:]
		}
		if virtualDirectory[0] == '/' || virtualDirectory[0] == os.PathSeparator { // second check for /something, (or ./something if we had dot on 0 it will be removed
			virtualDirectory = virtualDirectory[1:]
		}
	}
	var templateErr error

	s.mu.Lock()
	defer s.mu.Unlock()

	names := namesFn()
	for _, path := range names {
		if !strings.HasPrefix(path, virtualDirectory) {
			continue
		}
		ext := filepath.Ext(path)
		if ext == virtualExtension {

			rel, err := filepath.Rel(virtualDirectory, path)
			if err != nil {
				templateErr = err
				return err
			}

			buf, err := assetFn(path)
			if err != nil {
				templateErr = err
				return err
			}
			contents := string(buf)
			name := filepath.ToSlash(rel)

			tmpl, err := raymond.Parse(contents)
			if err != nil {
				templateErr = err
				return err
			}
			s.templateCache[name] = tmpl

		}
	}
	return templateErr
}

func (s *HandlebarsEngine) fromCache(relativeName string) *raymond.Template {
	s.mu.Lock()
	tmpl, ok := s.templateCache[relativeName]
	if ok {
		s.mu.Unlock()
		return tmpl
	}
	s.mu.Unlock()
	return nil
}

func (s *HandlebarsEngine) executeTemplateBuf(name string, binding interface{}) (string, error) {
	if tmpl := s.fromCache(name); tmpl != nil {
		return tmpl.Exec(binding)
	}
	return "", nil
}

// ExecuteWriter executes a template and writes its result to the w writer.
func (s *HandlebarsEngine) ExecuteWriter(w io.Writer, filename string, layout string, bindingData interface{}) error {
	// reload the templates if reload configuration field is true
	if s.reload {
		if err := s.Load(); err != nil {
			return err
		}
	}

	isLayout := false
	layout = getLayout(layout, s.layout)
	renderFilename := filename

	if layout != "" {
		isLayout = true
		renderFilename = layout // the render becomes the layout, and the name is the partial.
	}

	if tmpl := s.fromCache(renderFilename); tmpl != nil {
		binding := bindingData
		if isLayout {
			var context map[string]interface{}
			if m, is := binding.(map[string]interface{}); is { //handlebars accepts maps,
				context = m
			} else {
				return fmt.Errorf("Please provide a map[string]interface{} type as the binding instead of the %#v", binding)
			}

			contents, err := s.executeTemplateBuf(filename, binding)
			if err != nil {
				return err
			}
			if context == nil {
				context = make(map[string]interface{}, 1)
			}
			// I'm implemented the {{ yield }} as with the rest of template engines, so this is not inneed for siris, but the user can do that manually if want
			// there is no performanrce different: raymond.RegisterPartialTemplate(name, tmpl)
			context["yield"] = raymond.SafeString(contents)
		}

		res, err := tmpl.Exec(binding)

		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, res)
		return err
	}

	return fmt.Errorf("template with name %s[original name = %s] doesn't exists in the dir", renderFilename, filename)
}

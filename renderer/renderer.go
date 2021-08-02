/***
 * renderer is responsible for rendering data im the specific templates.
 */
package renderer

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/Masterminds/sprig/v3"

	"github.com/alexsergivan/mybooks/flash"
	"github.com/alexsergivan/mybooks/resolvers"
	"github.com/labstack/echo/v4/middleware"

	"github.com/alexsergivan/mybooks/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const BaseTemplatePrefix = `base:`

const templateExtension = `.gohtml`

var mutex = sync.RWMutex{}

var TemplatesDir string = "views"
var LayoutDir string = TemplatesDir + "/layouts"
var ComponentsDir string = TemplatesDir + "/components"

var once sync.Once

var view *View

type View struct {
	Template    *template.Template
	LayoutFiles []string
	tpls        embed.FS
	mux         sync.Mutex
}

type RenderData struct {
	sync.RWMutex
	Data map[string]interface{}
}

func (rd *RenderData) AddBulkData(values map[string]interface{}) {
	rd.Lock()
	defer rd.Unlock()
	for key, value := range values {
		rd.Data[key] = value
	}
}

func (rd *RenderData) AddData(key string, value interface{}) {
	rd.Lock()
	defer rd.Unlock()
	rd.Data[key] = value
}

func (rd *RenderData) GetData() map[string]interface{} {
	rd.RLock()
	defer rd.RUnlock()
	d := make(map[string]interface{}, len(rd.Data))
	for k, v := range rd.Data {
		d[k] = v
	}
	return d
}

func (rd *RenderData) GetValue(key string) interface{} {
	rd.RLock()
	defer rd.RUnlock()
	if val, ok := rd.Data[key]; ok {
		var v = val
		return v
	}
	return ""
}

func getFuncMap() template.FuncMap {

	return template.FuncMap{
		"hasField": HasField,
		"args":     ArgsFn,
		"to5Stars": services.ConvertRateFrom100To5,
		"toEmoji":  services.ConvertRateFrom100ToEmoji,
		"Iterate": func(count int) []int {
			var i int
			var Items []int
			for i = 0; i < (count); i++ {
				Items = append(Items, i)
			}
			return Items
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

}

func NewView(tpls embed.FS) *View {
	once.Do(func() {
		view = &View{
			Template:    nil,
			LayoutFiles: layoutFiles(tpls),
			tpls:        tpls,
			mux:         sync.Mutex{},
		}
	})

	return view
}

func RemoveHTMLComments(content []byte) []byte {
	htmlcmt := regexp.MustCompile(`<!--[^>]*-->`)
	return htmlcmt.ReplaceAll(content, []byte(""))
}

func MinifyHTML(html []byte) (string, error) {
	// read line by line
	minifiedHTML := ""
	tagOpen := false
	scanner := bufio.NewScanner(bytes.NewReader(RemoveHTMLComments(html)))
	for scanner.Scan() {
		// all leading and trailing white space of each line are removed
		lineTrimmed := strings.TrimSpace(scanner.Text())
		minifiedHTML += lineTrimmed
		if strings.Contains(lineTrimmed, "<") {
			tagOpen = true
		}
		if strings.Contains(lineTrimmed, ">") {
			tagOpen = false
		}
		if len(lineTrimmed) > 0 {
			// in case of following trimmed line:
			// <div id="foo"
			if tagOpen {
				minifiedHTML += " "
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return string(html), err
	}

	return minifiedHTML, nil
}

func (v *View) compileTemplates(tmpl *template.Template, filenames []string, tpls embed.FS) (*template.Template, error) {
	wg := sync.WaitGroup{}

	for _, filename := range filenames {
		wg.Add(1)

		go func(filename string) {

			b, err := tpls.ReadFile(filename)
			if err != nil {
				log.Print(err)
			}

			minifiedHTML, minifyErr := MinifyHTML(b)
			if minifyErr != nil {
				log.Printf("Error in Scanner during minifying the template", err)
			}
			v.mux.Lock()
			defer v.mux.Unlock()
			_, err = tmpl.Parse(minifiedHTML)
			if err != nil {
				log.Print(err)
			}
			wg.Done()
		}(filename)

	}
	wg.Wait()
	return tmpl, nil
}

// Render renders a component
func (v *View) Render(w io.Writer, componentName string, data interface{}, c echo.Context) error {
	templateName := `main`
	if strings.HasPrefix(componentName, BaseTemplatePrefix) {

		compSlice := strings.Split(componentName, "---")
		templateName = strings.TrimPrefix(compSlice[0], BaseTemplatePrefix)
		if len(compSlice) > 0 {
			componentName = compSlice[1]
		}
	}
	componentsFiles, err := componentsFiles(componentName, v.tpls)
	if err != nil {
		return err
	}

	files := append(v.LayoutFiles, componentsFiles...)

	templ := template.Must(v.compileTemplates(template.New("").Funcs(getFuncMap()).Funcs(sprig.FuncMap()), files, v.tpls))

	messageTypes := make(map[string][]string)
	mutex.Lock()
	for _, messageType := range flash.GetMessageTypes() {
		message, _ := flash.GetFlashMessage(c, messageType)
		messageTypes[messageType] = message
	}
	mutex.Unlock()

	rd := &RenderData{
		Data: map[string]interface{}{},
	}
	rd.AddBulkData(data.(map[string]interface{}))
	rd.AddBulkData(map[string]interface{}{
		"reverse":    c.Echo().Reverse,
		"activePath": c.Request().URL.Path,
		"user":       resolvers.GetCurrentUser(c),
		"csrf":       c.Get(middleware.DefaultCSRFConfig.ContextKey),
		"messages":   messageTypes,
	})

	return templ.ExecuteTemplate(w, templateName, rd.GetData())
}

// Gets list of all template files needed for the layout.
func layoutFiles(tpls embed.FS) []string {
	templates, err := GetFilenames(LayoutDir, templateExtension, tpls)
	if err != nil {
		log.Error(err)
	}

	return templates
}

func GetFilenames(rootDir, extension string, tpls embed.FS) ([]string, error) {
	var filenames []string
	err := fs.WalkDir(tpls, rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, extension) {
			filenames = append(filenames, path)
		}
		return nil
	})
	//err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
	//	if err != nil {
	//		return err
	//	}
	//
	//	if strings.HasSuffix(path, extension) {
	//		filenames = append(filenames, path)
	//	}
	//	return nil
	//})

	return filenames, err
}

// Gets list of all template files needed for the component.
// Example component name: places--city-connection.
func componentsFiles(componentName string, tpls embed.FS) ([]string, error) {
	componentNameSlice := strings.Split(componentName, "--")
	var compFiles []string
	path := ComponentsDir
	for _, compFolder := range componentNameSlice {
		path += "/" + compFolder
		//files, err := filepath.Glob(path + "/*" + templateExtension)
		files, err := fs.Glob(tpls, path+"/*"+templateExtension)

		if err != nil {
			return nil, err
		}
		compFiles = append(compFiles, files...)
	}

	return compFiles, nil
}

func HasField(v interface{}, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	return rv.FieldByName(name).IsValid()
}

// ArgsFn gives option to pass several number of arguments in template
func ArgsFn(kvs ...interface{}) (map[string]interface{}, error) {
	if len(kvs)%2 != 0 {
		return nil, errors.New("args requires even number of arguments")
	}
	m := make(map[string]interface{})
	for i := 0; i < len(kvs); i += 2 {
		s, ok := kvs[i].(string)
		if !ok {
			return nil, errors.New("even args to args must be strings")
		}
		m[s] = kvs[i+1]
	}
	return m, nil
}

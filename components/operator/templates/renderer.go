package templates

import (
	"embed"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

//go:embed all:*
var EFS embed.FS

func Render(name string, data *Data) ([]*unstructured.Unstructured, error) {
	rendered, err := renderStr(name, data)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(rendered))

	resList := &ResourceList{}
	if err := yaml.Unmarshal([]byte(rendered), resList); err != nil {
		return nil, err
	}

	return resList.Items, nil
}

func renderStr(name string, data *Data) (string, error) {
	tpl := template.New("list.tpl").Option("missingkey=zero")
	initFuncs(tpl, data)

	if _, err := tpl.ParseFS(EFS, "helpers.tpl", name+"/*"); err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return strings.ReplaceAll(buf.String(), "<no value>", ""), nil
}

func initFuncs(tpl *template.Template, data *Data) {
	funcMap := sprig.FuncMap()

	funcMap["include"] = func(name string, data interface{}) (string, error) {
		var buf strings.Builder
		err := tpl.ExecuteTemplate(&buf, name, data)
		return buf.String(), err
	}

	funcMap["toYaml"] = func(v any) string {
		data, err := yaml.Marshal(v)
		if err != nil {
			return ""
		}
		s := strings.TrimSuffix(string(data), "\n")
		if s == "null" {
			s = ""
		}
		return s
	}

	funcMap["file"] = func(name string) string {
		b, err := EFS.ReadFile(name)
		if err != nil {
			return ""
		}
		return string(b)
	}

	funcMap["namespace"] = data.Namespace
	funcMap["instanceName"] = data.InstanceName
	funcMap["platformName"] = data.PlatformName
	funcMap["appName"] = data.AppName
	funcMap["componentName"] = data.ComponentName

	tpl.Funcs(funcMap)
}

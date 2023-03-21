package utils

import (
	"fmt"
	"strings"
	"text/template"
)

func RenderingTemplate[T any](templateName, templateString string, data T) (s string, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("rendering template error: %v", r)
		}
	}()

	// prevent <no value>
	//parse, err := template.New(templateName).Option("missingkey=zero").Parse(templateString)
	parse, err := template.New(templateName).Parse(templateString)
	if err != nil {
		return "", err
	}
	builder := strings.Builder{}
	err = parse.Execute(&builder, data)
	if err != nil {
		return "", err
	}
	//return strings.ReplaceAll(builder.String(), "<no value>", ""), nil
	return builder.String(), nil
}

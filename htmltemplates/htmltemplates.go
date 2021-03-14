package htmltemplates

import (
	"fa-middleware/config"

	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
)

func GetFileAsStr(fileName string) (result string, err error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf(
			"error reading file %v: %v",
			fileName,
			err.Error(),
		)
	}
	return string(data), nil
}

func GetPaymentTemplate(conf config.Config /* user fusionauth.User */) (string, error) {
	templatestr, err := GetFileAsStr("templates/payment.html")
	if err != nil {
		return "", fmt.Errorf("failed to get file str: %v", err.Error())
	}

	t, err := template.New("payment").Parse(templatestr)
	if err != nil {
		return "", fmt.Errorf("failed to get html template: %v", err.Error())
	}

	data := struct {
		Title string
		Conf  config.Config
		// UserID   string
		// Username string
	}{
		Title: "Make a payment",
		Conf:  conf,
		// UserID:   user.Id,
		// Username: user.Username,
	}

	var executedTemplate bytes.Buffer
	err = t.Execute(&executedTemplate, data)
	if err != nil {
		return "", fmt.Errorf("failed to apply payment template: %v", err.Error())
	}

	return executedTemplate.String(), nil
}

// GetTemplateByName groups together the templates into a single function.
// templateName is the file name but not the whole path, such as "payment.html"
func GetTemplateByName(conf config.Config, user fusionauth.User, templateName string) (string, error) {
	templatePath := fmt.Sprintf(
		"templates/%v",
		templateName,
	)
	templatestr, err := GetFileAsStr(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to get file %v as str: %v", templatePath, err.Error())
	}

	t, err := template.New(templateName).Parse(templatestr)
	if err != nil {
		return "", fmt.Errorf("failed to get html template %v: %v", templateName, err.Error())
	}

	data := struct {
		Conf config.Config
		User fusionauth.User
	}{
		Conf: conf,
		User: user,
	}

	var executedTemplate bytes.Buffer
	err = t.Execute(&executedTemplate, data)
	if err != nil {
		return "", fmt.Errorf(
			"failed to apply %v template: %v",
			templateName,
			err.Error(),
		)
	}

	return executedTemplate.String(), nil
}

func GetLoggedInTemplate(user fusionauth.User) (string, error) {
	templatestr, err := GetFileAsStr("templates/loggedin.html")
	if err != nil {
		return "", fmt.Errorf("failed to get file str: %v", err.Error())
	}

	t, err := template.New("loggedin").Parse(templatestr)
	if err != nil {
		return "", fmt.Errorf("failed to get html template: %v", err.Error())
	}

	data := struct {
		Title    string
		UserID   string
		Username string
	}{
		Title:    fmt.Sprintf("Hello, %v", user.FullName),
		UserID:   user.Id,
		Username: user.Username,
	}

	var executedTemplate bytes.Buffer
	err = t.Execute(&executedTemplate, data)
	if err != nil {
		return "", fmt.Errorf("failed to apply payment template: %v", err.Error())
	}

	return executedTemplate.String(), nil
}

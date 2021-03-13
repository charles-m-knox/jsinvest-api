package htmltemplates

import (
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

func GetPaymentTemplate( /* user fusionauth.User */ ) (string, error) {
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
		// UserID   string
		// Username string
	}{
		Title: "Make a payment",
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

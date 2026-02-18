package handler

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"slices"
	"strings"

	"code.cacheflow.internal/util/secrets"
)

type EmailRequestBody struct {
	Email    string            `json:"email"`
	Subject  string            `json:"subject"`
	Template string            `json:"template"`
	Data     map[string]string `json:"data"`
}

func SendEmail(body EmailRequestBody) error {
	// List of allowed templates
	templates := []string{
		"verify-create-account",
	}

	if !slices.Contains(templates, body.Template) {
		return errors.New("invalid template")
	}

	emailTemplate, err := os.ReadFile("templates/" + body.Template + ".html")
	if err != nil {
		return err
	}

	emailTemplateStr := string(emailTemplate)
	for key, value := range body.Data {
		placeholder := "{{" + key + "}}"
		emailTemplateStr = strings.ReplaceAll(emailTemplateStr, placeholder, value)
	}

	subject := ""
	switch body.Template {
	case "verify-create-account":
		subject = "CacheFlow - Verify your email"
		default:
		return errors.New("invalid template")
	}

	headers := make(map[string]string)
	headers["From"] = "cacheflow.inquiry@gmail.com"
	headers["To"] = body.Email
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + emailTemplateStr

	auth := smtp.PlainAuth(
		"",                        // Identity (this is used for authentication, ex: username)
		"cacheflow.inquiry@gmail.com", // Email (from)
		secrets.EmailSecretValue,  // Password (for email)
		"smtp.gmail.com",          // Host
	)

	err = smtp.SendMail(
		"smtp.gmail.com:587",      // Host
		auth,                      // Auth
		"cacheflow.inquiry@gmail.com", // Email (from)
		[]string{body.Email},      // Email (to)
		[]byte(message),           // Message
	)
	if err != nil {
		return err
	}

	return nil
}
package mailagent

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/smtp"
	"path"
)

type MailAgent struct {
	plainAuth smtp.Auth
	templates map[string]*template.Template
	sender    string
}

type Value map[string]interface{}

func NewMailAgent(username string, password string) *MailAgent {
	return &MailAgent{
		plainAuth: smtp.PlainAuth("", username, password, "smtp.yandex.ru"),
		templates: make(map[string]*template.Template),
		sender:    username,
	}
}

func (m *MailAgent) LoadTemplatesFromDir(dir string) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range fs {
		if !f.IsDir() {
			templ, err := template.ParseFiles(path.Join(dir, f.Name()))
			if err != nil {
				return err
			}
			m.templates[f.Name()] = templ
			fmt.Printf("[mailagent] template `%s` loaded\r\n", f.Name())
		}
	}
	return err
}

func (m *MailAgent) execute(tmpl *template.Template, val Value) (string, error) {
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, val); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (m *MailAgent) SendTemplate(addr string, tmplName string, val Value) error {
	tmpl, ok := m.templates[tmplName]
	if !ok {
		return errors.New("template undefined")
	}

	html, err := m.execute(tmpl, val)
	if err != nil {
		return err
	}
	return m.Send(addr, html)
}

func (m *MailAgent) Send(addr string, msg string) error {
	var buf bytes.Buffer
	buf.WriteString("To: <" + addr + ">\r\n")
	buf.WriteString("From: POS-Ninja <" + m.sender + ">\r\n")
	buf.WriteString("Subject: Сообщение от POS-Ninja\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString(msg)

	return smtp.SendMail("smtp.yandex.ru:25", m.plainAuth, m.sender, []string{addr}, buf.Bytes())
}

package postman

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/smtp"
	"path"
)

type Postman struct {
	sender string

	plainAuth smtp.Auth
	templates map[string]*template.Template
}

type Value map[string]interface{}

func NewPosman(username string, password string) *Postman {
	return &Postman{
		sender: username,

		plainAuth: smtp.PlainAuth("", username, password, "smtp.yandex.ru"),
		templates: make(map[string]*template.Template),
	}
}

func (p *Postman) LoadTemplatesFromDir(dir string) error {
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
			p.templates[f.Name()] = templ
			fmt.Printf("[postman] template `%s` loaded\r\n", f.Name())
		}
	}
	return err
}

func (p *Postman) execute(tmpl *template.Template, val Value) (string, error) {
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, val); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (m *Postman) SendTemplate(addr string, tmplName string, val Value) error {
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

func (p *Postman) Send(addr string, msg string) error {
	var buf bytes.Buffer
	buf.WriteString("To: <" + addr + ">\r\n")
	buf.WriteString("From: POS-Ninja <" + p.sender + ">\r\n")
	buf.WriteString("Subject: Сообщение от POS-Ninja\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString(msg)

	return smtp.SendMail("smtp.yandex.ru:25", p.plainAuth, p.sender, []string{addr}, buf.Bytes())
}

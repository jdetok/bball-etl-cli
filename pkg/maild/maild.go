package maild

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type MIMEEmail struct {
	User string
	Pass string
	Host string
	Port string
	Addr string
	Subj string
	Body string
	Mesg string
	File string
	MlTo []string
	Auth smtp.Auth
}

type EmailAuth struct {
	User string
	Pass string
	Host string
	Port string
}

type EmailCnf struct {
	Auth *EmailAuth
	To   string
	Subj string
	Body string
	File string
}

func EmailLog(logf, userN, passN, hostN, portN string) error {
	a, err := NewEmailAuth(userN, passN, hostN, portN)
	if err != nil {
		return fmt.Errorf("failed to get email env: %v", err)
	}

	m := NewMIMEEmail(a,
		[]string{"jdekock17@gmail.com"},
		logf,
		"Go bball ETL log attached",
		"the Go bball ETL process ran. The log is attached.",
	)
	return m.SendMIMEEmail(logf)
}

func NewEmailAuth(userN, passN, hostN, portN string) (*EmailAuth, error) {
	a := &EmailAuth{}
	envVars := map[string]*string{
		hostN: &a.Host,
		portN: &a.Port,
		userN: &a.User,
		passN: &a.Pass,
	}
	for ev, v := range envVars {
		var tmp string
		if tmp = os.Getenv(ev); tmp == "" {
			return nil, fmt.Errorf("must set %s in .env", ev)
		}
		*v = tmp
	}
	return a, nil
}

func NewMIMEEmail(a *EmailAuth, to []string, file, subj, body string) *MIMEEmail {
	return &MIMEEmail{
		User: a.User,
		Pass: a.Pass,
		Host: a.Host,
		Port: a.Port,
		MlTo: to,
		Body: body,
		Subj: subj,
		File: file,
		Addr: fmt.Sprint(a.Host, ":", a.Port),
		Auth: smtp.PlainAuth("", a.User, a.Pass, a.Host),
	}
}

func (m *MIMEEmail) MakeMIMEMsg(fName string) error {
	atch, err := m.Attach(fName)
	if err != nil {
		return fmt.Errorf("error attaching file at %s: %v", fName, err)
	}
	bndry := "bndry-" + fName
	m.Mesg = strings.Join([]string{
		"From: " + m.User,
		"To: " + strings.Join(m.MlTo, ", "),
		"Subject: " + m.Subj,
		"MIME-Version: 1.0",
		"Content-Type: multipart/mixed; boundary=" + bndry,
		"",
		"--" + bndry,
		"Content-Type: text/plain; charset=utf-8",
		"Content-Transfer-Encoding: 7bit",
		"",
		m.Body,
		"",
		"--" + bndry,
		"Content-Type: application/octet-stream",
		fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"", m.File),
		"Content-Transfer-Encoding: base64",
		"",
		atch,
		"--" + bndry + "--",
		"",
	}, "\r\n")
	return nil
}

func (m *MIMEEmail) SendMIMEEmail(fName string) error {
	err := m.MakeMIMEMsg(fName)
	if err != nil {
		return fmt.Errorf("failed to create MIME msg with file attached at %s: %v", fName, err)
	}

	err = smtp.SendMail(m.Addr, m.Auth, m.User, m.MlTo, []byte(m.Mesg))
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}
	return nil
}

func (m *MIMEEmail) Attach(fName string) (string, error) {
	m.File = fName
	// read file at fName as []byte
	f, err := os.ReadFile(m.File)
	if err != nil {
		return "", fmt.Errorf("error reading file at %s: %v", m.File, err)
	}

	// encode bytes to base64 string
	enc := base64.StdEncoding.EncodeToString(f)
	return SplitFileLines(76, enc, "\r\n"), nil
}

func SplitFileLines(cLen int, body, delim string) string {
	var b strings.Builder
	for len(body) > cLen {
		b.WriteString(body[:cLen])
		b.WriteString(delim)
		body = body[cLen:]
	}
	b.WriteString(body)
	b.WriteString(delim)
	return b.String()
}

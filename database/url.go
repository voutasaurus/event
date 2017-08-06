package database

import (
	"bytes"
	"html/template"
	"net/url"
)

const sampleURL = "postgres://user:password@host:12345/dbname?sslrootcert=/path/to/root.pem?sslmode=verify-full"

type URL struct {
	User string
	Pass string
	Host string
	Port string
	Name string
	Root string
	Mode string
}

var defaultURL = URL{
	User: "postgres",
	Pass: "password",
	Host: "127.0.0.1",
	Port: "5432",
	Name: "db",
	Root: "",
	Mode: "verify-full",
}

func (u *URL) withDefaults() *URL {
	var v URL
	v = *u
	if v.User == "" {
		v.User = defaultURL.User
	}
	if v.Pass == "" {
		v.Pass = defaultURL.Pass
	}
	if v.Host == "" {
		v.Host = defaultURL.Host
	}
	if v.Port == "" {
		v.Port = defaultURL.Port
	}
	if v.Name == "" {
		v.Name = defaultURL.Name
	}
	if v.Root == "" {
		v.Root = defaultURL.Root
	}
	if v.Mode == "" {
		v.Mode = defaultURL.Mode
	}
	return &v
}

func (u *URL) params() string {
	if u.Mode == "" && u.Root == "" {
		return ""
	}
	v := make(url.Values)
	if u.Mode != "" {
		v.Add("sslmode", u.Mode)
	}
	if u.Root != "" {
		v.Add("sslrootcert", u.Root)
	}
	return "?" + v.Encode()
}

var urlTemplate = template.Must(template.New("connect").Parse("postgres://{{.User}}:{{.Pass}}@{{.Host}}:{{.Port}}/{{.Name}}{{.params}}"))

func (u *URL) String() string {
	v := u.withDefaults()
	var buf bytes.Buffer
	urlTemplate.Execute(&buf, v)
	return buf.String()
}

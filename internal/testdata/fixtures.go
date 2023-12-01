package testdata

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"testing"

	"git.rob.mx/nidito/joao/internal/testdata/opconnect"
	opclient "git.rob.mx/nidito/joao/pkg/op-client"
	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
)

func MockOPConnect(t *testing.T) {
	t.Helper()
	opclient.ConnectClientFactory = func(host, token, userAgent string) connect.Client {
		return &opconnect.Client{}
	}
	client := opclient.NewConnect("", "")
	opclient.Use(client)
	opconnect.Clear()
}

func FromProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	wd, _ := os.Getwd()
	return wd
}

func TempDir(t *testing.T, name string) string {
	newDir, err := os.MkdirTemp("", name+"-*")
	if err != nil {
		t.Fatalf("could not create tempdir")
	}
	return newDir
}

func YAML(name string) string {
	return path.Join(FromProjectRoot(), "testdata", fmt.Sprintf("%s.yaml", name))
}

func copyFile(in, out string) error {
	src, err := os.Open(in)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(out)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func TempYAML(t *testing.T, name string) (string, func()) {
	root := TempDir(t, "temp-yaml")
	path := fmt.Sprintf("%s/%s.yaml", root, name)
	if err := copyFile(YAML(name), path); err != nil {
		t.Fatalf("could not create copy of fixture %s.yaml", name)
	}
	return path, func() { os.Remove(path) }
}

func EnableDebugLogging() {
	logrus.SetLevel(logrus.DebugLevel)
}

func NewTestConfig(title string) *onepassword.Item {
	return &onepassword.Item{
		Title:    title,
		Vault:    onepassword.ItemVault{ID: "example"},
		Category: "PASSWORD",
		Sections: []*onepassword.ItemSection{
			{ID: "~annotations", Label: "~annotations"},
			{ID: "nested", Label: "nested"},
			{ID: "list", Label: "list"},
		},
		Fields: []*onepassword.ItemField{
			{
				ID:      "password",
				Type:    "CONCEALED",
				Purpose: "PASSWORD",
				Label:   "password",
				Value:   "8b23de7705b79b73d9f75b120651bc162859e45a732b764362feaefc882eab5d",
			},
			{
				ID:      "notesPlain",
				Type:    "STRING",
				Purpose: "NOTES",
				Label:   "notesPlain",
				Value:   "flushed by joao",
			},
			{
				ID:      "~annotations.int",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "int",
				Value:   "int",
			},
			{
				ID:    "int",
				Type:  "STRING",
				Label: "int",
				Value: "1",
			},
			{
				ID:    "string",
				Type:  "STRING",
				Label: "string",
				Value: "pato",
			},
			{
				ID:      "~annotations.bool",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "bool",
				Value:   "bool",
			},
			{
				ID:    "bool",
				Type:  "STRING",
				Label: "bool",
				Value: "false",
			},
			{
				ID:      "~annotations.secret",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "secret",
				Value:   "secret",
			},
			{
				ID:    "secret",
				Type:  "CONCEALED",
				Label: "secret",
				Value: "very secret",
			},
			{
				ID:      "~annotations.nested.int",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.int",
				Value:   "int",
			},
			{
				ID:      "nested.int",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Type:    "STRING",
				Label:   "int",
				Value:   "1",
			},
			{
				ID:      "~annotations.nested.bool",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.bool",
				Value:   "bool",
			},
			{
				ID:      "nested.bool",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Type:    "STRING",
				Label:   "bool",
				Value:   "true",
			},
			{
				ID:      "~annotations.nested.list.0",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.list.0",
				Value:   "int",
			},
			{
				ID:      "nested.list.0",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Type:    "STRING",
				Label:   "list.0",
				Value:   "1",
			},
			{
				ID:      "~annotations.nested.list.1",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.list.1",
				Value:   "int",
			},
			{
				ID:      "nested.list.1",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Type:    "STRING",
				Label:   "list.1",
				Value:   "2",
			},
			{
				ID:      "~annotations.nested.list.2",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.list.2",
				Value:   "int",
			},
			{
				ID:      "nested.list.2",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Type:    "STRING",
				Label:   "list.2",
				Value:   "3",
			},
			{
				ID:      "~annotations.nested.secret",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.secret",
				Value:   "secret",
			},
			{
				ID:      "nested.secret",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Type:    "CONCEALED",
				Label:   "secret",
				Value:   "very secret",
			},
			{
				ID:      "~annotations.nested.second_secret",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.second_secret",
				Value:   "secret",
			},
			{
				ID:      "nested.second_secret",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Type:    "CONCEALED",
				Label:   "second_secret",
				Value:   "very secret",
			},
			{
				ID:      "nested.string",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Type:    "STRING",
				Label:   "string",
				Value:   "quem",
			},
			{
				ID:      "list.0",
				Section: &onepassword.ItemSection{ID: "list", Label: "list"},
				Type:    "STRING",
				Label:   "0",
				Value:   "one",
			},
			{
				ID:      "list.1",
				Section: &onepassword.ItemSection{ID: "list", Label: "list"},
				Type:    "STRING",
				Label:   "1",
				Value:   "two",
			},
			{
				ID:      "list.2",
				Section: &onepassword.ItemSection{ID: "list", Label: "list"},
				Type:    "STRING",
				Label:   "2",
				Value:   "three",
			},
		},
	}
}

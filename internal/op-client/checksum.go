package opclient

import (
	"fmt"
	"sort"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"golang.org/x/crypto/blake2b"
)

func Checksum(fields []*op.ItemField) string {
	newHash, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	df := []string{}
	for _, field := range fields {
		if field.ID == "password" || field.ID == "notesPlain" || (field.Section != nil && field.Section.ID == "~annotations") {
			continue
		}
		label := field.Label
		if field.Section != nil && field.Section.ID != "" {
			label = field.Section.ID + "." + label
		}
		df = append(df, label+field.Value)
	}
	sort.Strings(df)
	newHash.Write([]byte(strings.Join(df, "")))
	checksum := newHash.Sum(nil)
	return fmt.Sprintf("%x", checksum)
}

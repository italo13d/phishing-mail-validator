package preprocess

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jaytaylor/html2text"
)

// MaxLen define o corte (3k chars por padrão)
const MaxLen = 3000

var quoteRE = regexp.MustCompile(`(?s)(On .*? wrote:|-----Mensagem.*?-----).*`)

func Clean(subj, from, to, html string) string {
	txt, _ := html2text.FromString(html, html2text.Options{PrettyTables: false})
	txt = quoteRE.ReplaceAllString(txt, "")
	if len(txt) > MaxLen {
		txt = txt[:MaxLen]
	}
	return fmt.Sprintf("Subject: %s\nFrom: %s\nTo: %s\n\n%s",
		subj, from, to, strings.TrimSpace(txt))
}

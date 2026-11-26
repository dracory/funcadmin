package shared

import (
	"strconv"

	"github.com/dracory/hb"
)

// ErrorPopup returns a Notiflix error notification tag.
func ErrorPopup(errorMessage string) hb.TagInterface {
	return hb.Script(`Notiflix.Notify.failure('` + escapeJS(errorMessage) + `');`)
}

// SuccessPopup returns a Notiflix success notification tag.
func SuccessPopup(successMessage string) hb.TagInterface {
	return hb.Script(`Notiflix.Notify.success('` + escapeJS(successMessage) + `');`)
}

// SuccessPopupWithRedirect returns a Notiflix success notification tag
// with an optional redirect. If redirectUrl is empty, no redirect is
// configured.
func SuccessPopupWithRedirect(successMessage string, redirectUrl string, redirectSeconds int) hb.TagInterface {
	if redirectUrl != "" {
		return hb.Script(`Notiflix.Notify.success('` + escapeJS(successMessage) + `'); setTimeout(function(){ window.location.href = '` + escapeJS(redirectUrl) + `'; }, ` + strconv.Itoa(redirectSeconds*1000) + `);`)
	}

	return hb.Script(`Notiflix.Notify.success('` + escapeJS(successMessage) + `');`)
}

// escapeJS escapes a string for safe embedding in a single-quoted JS
// string literal.
func escapeJS(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\':
			out = append(out, '\\', '\\')
		case '\'':
			out = append(out, '\\', '\'')
		case '\n':
			out = append(out, '\\', 'n')
		case '\r':
			out = append(out, '\\', 'r')
		case '\t':
			out = append(out, '\\', 't')
		default:
			out = append(out, c)
		}
	}
	return string(out)
}

// formatInt64 is a helper used by header.go
func formatInt64(n int64) string {
	if n < 0 {
		return "?"
	}
	return strconv.FormatInt(n, 10)
}

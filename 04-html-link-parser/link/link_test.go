package link

import (
	"cmp"
	"slices"
	"strings"
	"testing"
)

func sortLinks(l []Link) {
	slices.SortFunc(l, func(a, b Link) int {
		return cmp.Compare(a.Href, b.Href)
	})
}

func compareLinks(t *testing.T, got, want []Link) {
	t.Helper()
	sortLinks(got)
	sortLinks(want)

	if len(got) != len(want) {
		t.Errorf("Expected %d links, but got %d links instead", len(want), len(got))
	}

	for i := range len(got) {
		if got[i] != want[i] {
			t.Errorf("Element #%d mismatch: expected %v, got %v", i, want, got)
		}
	}

}

func TestSingleLink(t *testing.T) {
	html := `<a href="/other-page">A link to another page</a>`
	r := strings.NewReader(html)
	got, err := ParseHtml(r)
	want := []Link{
		{Href: "/other-page", Text: "A link to another page"},
	}
	if err != nil {
		t.Errorf("unexpected error parsing html: %s", err)
	}
	compareLinks(t, got, want)

}

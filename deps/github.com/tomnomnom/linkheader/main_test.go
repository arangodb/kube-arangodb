package linkheader

import "testing"

func TestSimple(t *testing.T) {
	// Test case stolen from https://github.com/thlorenz/parse-link-header :)
	header := "<https://api.github.com/user/9287/repos?page=3&per_page=100>; rel=\"next\", " +
		"<https://api.github.com/user/9287/repos?page=1&per_page=100>; rel=\"prev\"; pet=\"cat\", " +
		"<https://api.github.com/user/9287/repos?page=5&per_page=100>; rel=\"last\""

	links := Parse(header)

	if len(links) != 3 {
		t.Errorf("Should have been 3 links returned, got %d", len(links))
	}

	if links[0].URL != "https://api.github.com/user/9287/repos?page=3&per_page=100" {
		t.Errorf("First link should have URL 'https://api.github.com/user/9287/repos?page=3&per_page=100'")
	}

	if links[0].Rel != "next" {
		t.Errorf("First link should have rel=\"next\"")
	}

	if len(links[0].Params) != 0 {
		t.Errorf("First link should have exactly 0 params, but has %d", len(links[0].Params))
	}

	if len(links[1].Params) != 1 {
		t.Errorf("Second link should have exactly 1 params, but has %d", len(links[1].Params))
	}

	if links[1].Params["pet"] != "cat" {
		t.Errorf("Second link's 'pet' param should be 'cat', but was %s", links[1].Params["pet"])
	}

}

func TestEmpty(t *testing.T) {
	links := Parse("")
	if links != nil {
		t.Errorf("Return value should be nil, but was %d", len(links))
	}
}

// Although not often seen in the wild, the grammar in RFC 5988 suggests that it's
// valid for a link header to have nothing but a URL.
func TestNoRel(t *testing.T) {
	links := Parse("<http://example.com>")

	if len(links) != 1 {
		t.Fatalf("Length of links should be 1, but was %d", len(links))
	}

	if links[0].URL != "http://example.com" {
		t.Errorf("URL should be http://example.com, but was %s", links[0].URL)
	}
}

func TestLinkMethods(t *testing.T) {
	header := "<https://api.github.com/user/9287/repos?page=1&per_page=100>; rel=\"prev\"; pet=\"cat\""
	links := Parse(header)
	link := links[0]

	if link.HasParam("foo") {
		t.Errorf("Link should not have param 'foo'")
	}

	val := link.Param("pet")
	if val != "cat" {
		t.Errorf("Link should have param pet=\"cat\"")
	}

	val = link.Param("foo")
	if val != "" {
		t.Errorf("Link should not have value for param 'foo'")
	}

}

func TestLinksMethods(t *testing.T) {
	header := "<https://api.github.com/user/9287/repos?page=3&per_page=100>; rel=\"next\", " +
		"<https://api.github.com/user/9287/repos?page=1&per_page=100>; rel=\"stylesheet\"; pet=\"cat\", " +
		"<https://api.github.com/user/9287/repos?page=5&per_page=100>; rel=\"stylesheet\""

	links := Parse(header)

	filtered := links.FilterByRel("next")

	if filtered[0].URL != "https://api.github.com/user/9287/repos?page=3&per_page=100" {
		t.Errorf("URL did not match expected")
	}

	filtered = links.FilterByRel("stylesheet")
	if len(filtered) != 2 {
		t.Errorf("Filter for stylesheet should yield 2 results but got %d", len(filtered))
	}

	filtered = links.FilterByRel("notarel")
	if len(filtered) != 0 {
		t.Errorf("Filter by non-existant rel should yeild no results")
	}

}

func TestParseMultiple(t *testing.T) {
	headers := []string{
		"<https://api.github.com/user/58276/repos?page=2>; rel=\"next\"",
		"<https://api.github.com/user/58276/repos?page=2>; rel=\"last\"",
	}

	links := ParseMultiple(headers)

	if len(links) != 2 {
		t.Errorf("Should have returned 2 links")
	}
}

func TestLinkToString(t *testing.T) {
	l := Link{
		URL: "http://example.com/page/2",
		Rel: "next",
		Params: map[string]string{
			"foo": "bar",
		},
	}

	have := l.String()

	parsed := Parse(have)

	if len(parsed) != 1 {
		t.Errorf("Expected only 1 link")
	}

	if parsed[0].URL != l.URL {
		t.Errorf("Re-parsed link header should have matching URL, but has `%s`", parsed[0].URL)
	}

	if parsed[0].Rel != l.Rel {
		t.Errorf("Re-parsed link header should have matching rel, but has `%s`", parsed[0].Rel)
	}

	if parsed[0].Param("foo") != "bar" {
		t.Errorf("Re-parsed link header should have foo=\"bar\" but doesn't")
	}
}

func TestLinksToString(t *testing.T) {
	ls := Links{
		{URL: "http://example.com/page/3", Rel: "next"},
		{URL: "http://example.com/page/1", Rel: "last"},
	}

	have := ls.String()

	want := "<http://example.com/page/3>; rel=\"next\", <http://example.com/page/1>; rel=\"last\""

	if have != want {
		t.Errorf("Want `%s`, have `%s`", want, have)
	}
}

func BenchmarkParse(b *testing.B) {

	header := "<https://api.github.com/user/9287/repos?page=3&per_page=100>; rel=\"next\", " +
		"<https://api.github.com/user/9287/repos?page=1&per_page=100>; rel=\"prev\"; pet=\"cat\", " +
		"<https://api.github.com/user/9287/repos?page=5&per_page=100>; rel=\"last\""

	for i := 0; i < b.N; i++ {
		_ = Parse(header)
	}
}

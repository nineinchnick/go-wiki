package main

import (
	"testing"
)

func assertEqPage(t *testing.T, result, expected *Page) {
	assertEq(t, result.Title, expected.Title)
	assertEq(t, string(result.Body), string(expected.Body))
}

func assertEq(t *testing.T, result, expected interface{}) {
	if result != expected {
		t.Errorf("Expecting %s, got %s", expected, result)
	}
}

func assertNe(t *testing.T, result, expected interface{}) {
	if result == expected {
		t.Errorf("Not expecting %s", result)
	}
}

func TestLoadPage(t *testing.T) {
	result, err := loadPage("testdata", "example")

	expected := &Page{Title: "example", Body: []byte("content\n")}
	assertEq(t, err, nil)
	assertEqPage(t, result, expected)
}

func TestLoadPageError(t *testing.T) {
	_, err := loadPage("testdata", "invalid")

	assertNe(t, err, nil)
}

func TestSavePage(t *testing.T) {
	example := Page{Title: "other_example", Body: []byte("other content\n")}

	err := example.save("testdata")
	assertEq(t, err, nil)

	result, _ := loadPage("testdata", "other_example")
	assertEqPage(t, result, &example)
}

func TestSavePageError(t *testing.T) {
	example := Page{Title: "other_example", Body: []byte("other content\n")}

	err := example.save("invalid")

	assertNe(t, err, nil)
}

func TestLinkPage(t *testing.T) {
	input := []byte("Some text with a [link].")

	result := linkPages(input, "http://localhost:8080/view/")

	expected := "Some text with a <a href=\"http://localhost:8080/view/link\">link</a>."
	assertEq(t, string(result), expected)
}

func TestAutoIndex(t *testing.T) {
	result := autoIndex("testdata", "http://localhost:8080/view/")

	expected := `<ul><li><a href="http://localhost:8080/view/example">example</a></li><li><a href="http://localhost:8080/view/other_example">other_example</a></li></ul>`
	assertEq(t, string(result), expected)

	result = autoIndex("invalid", "http://localhost:8080/view/")

	expected = "No pages"
	assertEq(t, string(result), expected)
}

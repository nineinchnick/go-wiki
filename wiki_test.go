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
	// When
	result, err := loadPage("testdata", "example")

	// Then
	expected := &Page{Title: "example", Body: []byte("content\n")}
	assertEq(t, err, nil)
	assertEqPage(t, result, expected)
}

func TestLoadPageError(t *testing.T) {
	// When
	_, err := loadPage("testdata", "invalid")

	// Then
	assertNe(t, err, nil)
}

func TestSavePage(t *testing.T) {
	// Having
	example := Page{Title: "other_example", Body: []byte("other content\n")}

	// When
	err := example.save("testdata")
	assertEq(t, err, nil)

	// Then
	result, _ := loadPage("testdata", "other_example")
	assertEqPage(t, result, &example)
}

func TestSavePageError(t *testing.T) {
	// Having
	example := Page{Title: "other_example", Body: []byte("other content\n")}

	// When
	err := example.save("invalid")

	// Then
	assertNe(t, err, nil)
}

func TestLinkPage(t *testing.T) {
	// Having
	input := []byte("Some text with a [link].")

	// When
	result := linkPages(input, "http://localhost:8080/view/")

	// Then
	expected := "Some text with a <a href=\"http://localhost:8080/view/link\">link</a>."
	assertEq(t, string(result), expected)
}

func TestAutoIndex(t *testing.T) {
	// When
	result := autoIndex("testdata", "http://localhost:8080/view/")

	// Then
	expected := `<ul><li><a href="http://localhost:8080/view/example">example</a></li><li><a href="http://localhost:8080/view/other_example">other_example</a></li></ul>`
	assertEq(t, string(result), expected)

	// When
	result = autoIndex("invalid", "http://localhost:8080/view/")

	// Then
	expected = "No pages"
	assertEq(t, string(result), expected)
}

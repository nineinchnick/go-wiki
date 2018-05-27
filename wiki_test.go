package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
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

func assertNotContains(t *testing.T, result, expected string) {
	if strings.Contains(result, expected) {
		t.Errorf("String %s contains %s", result, expected)
	}
}

func assertContains(t *testing.T, result, expected string) {
	if !strings.Contains(result, expected) {
		t.Errorf("String %s does not contains %s", result, expected)
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

	err := example.Save("testdata")
	assertEq(t, err, nil)

	result, _ := loadPage("testdata", "other_example")
	assertEqPage(t, result, &example)
}

func TestSavePageError(t *testing.T) {
	example := Page{Title: "other_example", Body: []byte("other content\n")}

	err := example.Save("invalid")

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

func TestGetFiles(t *testing.T) {
	result := getFiles("testdata/*")

	assertEq(t, len(result), 5)
}

func TestGetTemplateFiles(t *testing.T) {
	result := getTemplateFiles("testdata")

	assertEq(t, len(result), 1)
	assertEq(t, result["template.tpl"][0], "testdata/layout.html")
	assertEq(t, result["template.tpl"][1], "testdata/template.tpl")
}

func TestViewHandler(t *testing.T) {
	templates = loadTemplates()
	request, _ := http.NewRequest("GET", "http://localhost:8080/view/example", nil)
	w := httptest.NewRecorder()
	viewHandler(w, request, "testdata", "example")
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assertEq(t, resp.StatusCode, http.StatusOK)
	assertNotContains(t, string(body), "form")
	assertContains(t, string(body), "content\n")
}

func TestViewHandlerNew(t *testing.T) {
	templates = loadTemplates()
	request, _ := http.NewRequest("GET", "http://localhost:8080/view/invalid", nil)
	w := httptest.NewRecorder()
	viewHandler(w, request, "testdata", "invalid")
	resp := w.Result()
	assertEq(t, resp.StatusCode, http.StatusFound)
	assertEq(t, resp.Header["Location"][0], "/edit/invalid")
}

func TestEditHandler(t *testing.T) {
	templates = loadTemplates()
	request, _ := http.NewRequest("GET", "http://localhost:8080/edit/example", nil)
	w := httptest.NewRecorder()
	editHandler(w, request, "testdata", "example")
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assertEq(t, resp.StatusCode, http.StatusOK)
	assertContains(t, string(body), "form")
	assertContains(t, string(body), "content\n")
}

func TestEditHandlerNew(t *testing.T) {
	templates = loadTemplates()
	request, _ := http.NewRequest("GET", "http://localhost:8080/edit/invalid", nil)
	w := httptest.NewRecorder()
	editHandler(w, request, "testdata", "invalid")
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assertEq(t, resp.StatusCode, http.StatusOK)
	assertContains(t, string(body), "form")
	assertNotContains(t, string(body), "content\n")
}

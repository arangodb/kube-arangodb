package helper

import (
	"html"
	"log"
)

// LinkTag make css link tag from url
func LinkTag(url string) string {
	return `<link type="text/css" rel="stylesheet" href="` + html.EscapeString(url) + `"></link>`
}

// ScriptTag make js script tag from url
func ScriptTag(url string) string {
	return `<script type="text/javascript" src="` + html.EscapeString(url) + `"></script>`
}

// AssetTag make js or css tag from url
func AssetTag(kind, url string) string {
	var buf string
	if kind == "css" {
		buf = LinkTag(url)
	} else if kind == "js" {
		buf = ScriptTag(url)
	} else {
		log.Println("go-webpack: unsupported asset kind: " + kind)
		buf = ""
	}
	return buf
}

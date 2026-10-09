package security

import (
	"errors"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const MaxRichTextBytes = 1 << 20

var allowedRichTextTags = map[string]bool{
	"p": true, "h2": true, "h3": true, "strong": true, "em": true,
	"ul": true, "ol": true, "li": true, "blockquote": true, "a": true, "br": true,
}

// SanitizeRichText keeps the small formatting vocabulary approved for content
// while removing executable elements, event attributes, and unsafe URLs.
func SanitizeRichText(input string) (string, error) {
	if len(input) > MaxRichTextBytes {
		return "", errors.New("rich text exceeds maximum size")
	}
	root, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", errors.New("invalid rich text")
	}
	var output strings.Builder
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if err := renderSafeNode(&output, child); err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(output.String()), nil
}

func renderSafeNode(output *strings.Builder, node *html.Node) error {
	switch node.Type {
	case html.TextNode:
		if parentIsDocument(node) {
			return nil
		}
		output.WriteString(html.EscapeString(node.Data))
	case html.ElementNode:
		tag := strings.ToLower(node.Data)
		if tag == "script" || tag == "style" || tag == "iframe" || tag == "svg" || tag == "math" {
			return nil
		}
		if !allowedRichTextTags[tag] {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if err := renderSafeNode(output, child); err != nil {
					return err
				}
			}
			return nil
		}
		output.WriteByte('<')
		output.WriteString(tag)
		if tag == "a" {
			for _, attr := range node.Attr {
				if strings.EqualFold(attr.Key, "href") && safeLink(attr.Val) {
					output.WriteString(` href="`)
					output.WriteString(html.EscapeString(strings.TrimSpace(attr.Val)))
					output.WriteByte('"')
					break
				}
			}
		}
		output.WriteByte('>')
		if tag != "br" {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if err := renderSafeNode(output, child); err != nil {
					return err
				}
			}
			output.WriteString("</")
			output.WriteString(tag)
			output.WriteByte('>')
		}
	case html.DocumentNode:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := renderSafeNode(output, child); err != nil {
				return err
			}
		}
	}
	return nil
}

func parentIsDocument(node *html.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode {
			return false
		}
	}
	return true
}

func safeLink(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "mailto":
		return parsed.Scheme != "mailto" || parsed.Opaque != ""
	default:
		return false
	}
}

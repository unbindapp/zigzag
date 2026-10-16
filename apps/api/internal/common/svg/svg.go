package svg

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
)

const (
	ContentType = "image/svg+xml"
	MaxBytes    = 1 << 20
)

var (
	blockedElements = map[string]string{
		"script":        "script elements",
		"foreignobject": "foreignObject elements",
		"iframe":        "iframe elements",
		"embed":         "embed elements",
		"object":        "object elements",
	}
	blockedValues       = []string{"javascript:", "data:text/html"}
	externalURLPrefixes = []string{"http:", "https:", "//"}
)

func Validate(data []byte) (int, int, error) {
	if len(data) > MaxBytes {
		return 0, 0, invalid("SVG images are limited to 1 MiB")
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	var root *xml.StartElement
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, 0, invalid("The SVG is not valid XML")
		}
		switch token := token.(type) {
		case xml.Directive:
			return 0, 0, invalid("SVG images cannot contain a DOCTYPE or entity declarations")
		case xml.ProcInst:
			if token.Target != "xml" {
				return 0, 0, invalid("SVG images cannot contain processing instructions")
			}
		case xml.StartElement:
			if root == nil {
				if !strings.EqualFold(token.Name.Local, "svg") {
					return 0, 0, invalid("The root element must be <svg>")
				}
				copied := token.Copy()
				root = &copied
			}
			if err := checkElement(token); err != nil {
				return 0, 0, err
			}
		}
	}
	if root == nil {
		return 0, 0, invalid("The root element must be <svg>")
	}
	width, height := size(root)
	return width, height, nil
}

func checkElement(element xml.StartElement) error {
	name := strings.ToLower(element.Name.Local)
	if what, blocked := blockedElements[name]; blocked {
		return invalid("SVG images cannot contain " + what)
	}
	for _, attr := range element.Attr {
		attrName := strings.ToLower(attr.Name.Local)
		if strings.HasPrefix(attrName, "on") {
			return invalid("SVG images cannot contain event handler attributes")
		}
		value := normalized(attr.Value)
		for _, blockedValue := range blockedValues {
			if strings.Contains(value, blockedValue) {
				return invalid("SVG images cannot contain javascript: or data:text/html values")
			}
		}
		if (name == "image" || name == "use") && (attrName == "href" || attrName == "src") && isExternalURL(value) {
			return invalid("SVG images cannot reference external images")
		}
	}
	return nil
}

func normalized(value string) string {
	return strings.Map(func(r rune) rune {
		if r <= ' ' || r == 0x7f {
			return -1
		}
		return r
	}, strings.ToLower(value))
}

func isExternalURL(value string) bool {
	for _, prefix := range externalURLPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func size(root *xml.StartElement) (int, int) {
	var widthAttr, heightAttr, viewBox string
	for _, attr := range root.Attr {
		if attr.Name.Space != "" {
			continue
		}
		switch attr.Name.Local {
		case "width":
			widthAttr = attr.Value
		case "height":
			heightAttr = attr.Value
		case "viewBox":
			viewBox = attr.Value
		}
	}
	width, height := length(widthAttr), length(heightAttr)
	if width > 0 && height > 0 {
		return width, height
	}
	fields := strings.FieldsFunc(viewBox, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r' })
	if len(fields) != 4 {
		return 0, 0
	}
	width, height = length(fields[2]), length(fields[3])
	if width <= 0 || height <= 0 {
		return 0, 0
	}
	return width, height
}

func length(value string) int {
	number, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(value), "px"), 64)
	if err != nil || !(number >= 1 && number <= math.MaxInt32) {
		return 0
	}
	return int(math.Round(number))
}

func invalid(message string) error {
	return errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, message)
}

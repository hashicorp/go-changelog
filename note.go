package changelog

import (
	"regexp"
	"sort"
	"strings"
)

type Note struct {
	Type  string
	Body  string
	Issue string
}

var textInBodyREs = []*regexp.Regexp{
	regexp.MustCompile("(?ms)^```release-note\n(?P<note>.+)\n```"),
	regexp.MustCompile("(?ms)^```releasenote\n(?P<note>.+)\n```"),
	regexp.MustCompile("(?ms)^```release-note:(?P<type>[^\n]*)\n?(?P<note>.*)\n?```"),
	regexp.MustCompile("(?ms)^```releasenote:(?P<type>[^\n]*)\n?(?P<note>.*)\n?```"),
}

func NotesFromEntry(entry Entry) []Note {
	var res []Note
	for _, re := range textInBodyREs {
		matches := re.FindAllStringSubmatch(entry.Body, -1)
		if len(matches) == 0 {
			continue
		}

		for _, match := range matches {
			note := ""
			typ := ""
			for i, name := range re.SubexpNames() {
				switch name {
				case "note":
					note = match[i]
				case "type":
					typ = match[i]
				}
				if note != "" && typ != "" {
					break
				}
			}

			note = strings.TrimSpace(note)
			typ = strings.TrimSpace(typ)

			if note == "" && typ == "" {
				continue
			}

			res = append(res, Note{
				Type:  typ,
				Body:  note,
				Issue: entry.Issue,
			})
		}
	}
	sort.Slice(res, SortNotes(res))
	return res
}

func SortNotes(res []Note) func(i, j int) bool {
	return func(i, j int) bool {
		if res[i].Type < res[j].Type {
			return true
		} else if res[j].Type < res[i].Type {
			return false
		} else if res[i].Issue < res[j].Issue {
			return true
		} else if res[j].Issue < res[i].Issue {
			return false
		} else if res[i].Body < res[j].Body {
			return true
		} else if res[j].Body < res[i].Body {
			return false
		}
		return false
	}
}

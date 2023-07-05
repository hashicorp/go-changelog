package parser

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
)

var (
	defaultSectionReFmt = `(?s)(?P<header>## %s[^\n]*)
(?P<body>.+?)
(?:## .+|$)`
	headerMatchName = "header"
	bodyMatchName   = "body"
)

type SectionParser struct {
	RegexpFormat string
	content      []byte
}

func NewSectionParser(r io.Reader) (*SectionParser, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &SectionParser{
		RegexpFormat: defaultSectionReFmt,
		content:      b,
	}, nil
}

type SectionRange struct {
	HeaderRange *ByteRange
	BodyRange   *ByteRange
}

type ByteRange struct {
	From, To int
}

func (p *SectionParser) regexpFormat() string {
	if p.RegexpFormat == "" {
		return defaultSectionReFmt
	}
	return p.RegexpFormat
}

func (p *SectionParser) regexp(v string) (*regexp.Regexp, error) {
	escapedVersion := regexp.QuoteMeta(v)
	return regexp.Compile(fmt.Sprintf(p.regexpFormat(), escapedVersion))
}

func (p *SectionParser) SectionRange(v string) (*SectionRange, error) {
	re, err := p.regexp(v)
	if err != nil {
		return nil, err
	}

	loc := re.FindSubmatchIndex(p.content)
	if loc == nil {
		return nil, &VersionNotFoundErr{v}
	}

	headerIdx, err := findSubexpIndexes(re, headerMatchName)
	if err != nil {
		return nil, err
	}

	bodyIdx, err := findSubexpIndexes(re, bodyMatchName)
	if err != nil {
		return nil, err
	}

	return &SectionRange{
		HeaderRange: &ByteRange{
			From: loc[headerIdx.from],
			To:   loc[headerIdx.to],
		},
		BodyRange: &ByteRange{
			From: loc[bodyIdx.from],
			To:   loc[bodyIdx.to],
		},
	}, nil
}

type index struct {
	from, to int
}

func findSubexpIndexes(re *regexp.Regexp, name string) (*index, error) {
	for i, seName := range re.SubexpNames() {
		if seName == name {
			from := i * 2
			return &index{from, from + 1}, nil
		}
	}

	return nil, fmt.Errorf("subexpression %q not found", name)
}

type Section struct {
	Header []byte
	Body   []byte
}

func (p *SectionParser) Section(v string) (*Section, error) {
	sr, err := p.SectionRange(v)
	if err != nil {
		return nil, err
	}

	headerRng := sr.HeaderRange
	bodyRng := sr.BodyRange

	return &Section{
		Header: bytes.TrimSpace(p.content[headerRng.From:headerRng.To]),
		Body:   bytes.TrimSpace(p.content[bodyRng.From:bodyRng.To]),
	}, nil
}

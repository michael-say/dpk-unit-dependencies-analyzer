/*
	Copyright (c) 2020 Michael Saigachenko
*/

package dpk

import (
	"regexp"
	"strings"
)

type parser struct {
	text string
	pos  int
}

func createParser(str string) *parser {
	return &parser{
		pos:  0,
		text: str,
	}
}

func (p *parser) hasMore(length int) bool {
	return (p.pos + length) <= len(p.text)
}

func (p *parser) accept(seq string) bool {
	if !p.hasMore(len(seq)) {
		return false
	}
	for i := 0; i < len(seq); i++ {
		if !strings.EqualFold(seq, p.text[p.pos:p.pos+len(seq)]) {
			return false
		}
	}
	p.pos += len(seq)
	return true
}

func (p *parser) nextch() int {
	if p.eof() {
		return -1
	}
	ch := p.text[p.pos]
	p.pos++
	return int(ch)
}

func (p *parser) eof() bool {
	return p.pos == len(p.text)
}

func (p *parser) readTillEol() string {
	var str strings.Builder
	for {
		next := p.nextch()
		if next == -1 || next == '\n' {
			return str.String()
		}
		str.WriteByte(byte(next))
	}
}

func (p *parser) nextLine() string {
	for {
		if p.accept("//") {
			p.readTillEol()
		}
		return p.readTillEol()
	}
}

func (p *parser) skipUntil(lineExpr string) {
	for {
		line := p.nextLine()
		re := regexp.MustCompile(lineExpr)
		if re.MatchString(line) || p.eof() {
			return
		}
	}
}

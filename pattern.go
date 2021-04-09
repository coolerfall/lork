// Copyright (c) 2019-2021 Vincent Cheung (coolingfall@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package slago

import (
	"bytes"
	"fmt"
	"strings"
	"text/scanner"
)

const (
	typeLiteral nodeType = iota + 1
	typeSingle
	typeComposite
)

type nodeType int

type node struct {
	_type   nodeType
	value   string
	options []string
	next    *node
	child   *node
}

type patternParser struct {
	pattern string

	head *node
	tail *node
}

// NewPatternParser creates a new instance of pattern parser.
func NewPatternParser(pattern string) *patternParser {
	return &patternParser{
		pattern: pattern,
	}
}

// Parse parses pattern to pattern node chain.
func (p *patternParser) Parse() (*node, error) {
	var buf = new(bytes.Buffer)
	var keywordStart bool
	var compositeStart bool
	var optionStart bool
	var bracketCount int
	var keyword string

	var s scanner.Scanner
	s.Init(strings.NewReader(p.pattern))
	s.Whitespace = 1<<'\t' | 1<<'\n' | 1<<'\r'

	for tk := s.Scan(); tk != scanner.EOF; tk = s.Scan() {
		value := s.TokenText()
		switch value {
		case "#":
			if !keywordStart {
				keywordStart = true
			} else {
				buf.WriteString(value)
			}
		case "(":
			if keywordStart && !compositeStart {
				compositeStart = true
			} else {
				buf.WriteString(value)
			}
			bracketCount++
		case ")":
			bracketCount--
			if bracketCount == 0 {
				child, err := NewPatternParser(buf.String()).Parse()
				if err != nil {
					return nil, err
				}
				buf.Reset()
				p.appendNode(&node{
					_type: typeComposite,
					value: keyword,
					child: child,
				})
				keyword = ""
				keywordStart = false
				compositeStart = false
			} else {
				buf.WriteString(value)
			}
		case "{":
			if compositeStart {
				buf.WriteString(value)
			} else {
				if keywordStart && len(keyword) != 0 {
					p.appendNode(&node{
						_type: typeSingle,
						value: keyword,
					})
					keyword = ""
					keywordStart = false
				}
				optionStart = true
			}
		case "}":
			if optionStart {
				opt := []string{buf.String()}
				buf.Reset()
				p.tail.options = opt
				optionStart = false
			} else {
				buf.WriteString(value)
			}
		case ",":
			if compositeStart {
				buf.WriteString(value)
			}
		default:
			if compositeStart || optionStart {
				buf.WriteString(value)
			} else if keywordStart {
				if len(keyword) == 0 {
					keyword = value
				} else {
					p.appendNode(&node{
						_type: typeSingle,
						value: keyword,
					})
					keyword = ""
					keywordStart = false

					p.appendNode(&node{
						_type: typeLiteral,
						value: value,
					})
				}
			} else {
				p.appendNode(&node{
					_type: typeLiteral,
					value: value,
				})
			}
		}
	}

	if len(keyword) != 0 {
		p.appendNode(&node{
			_type: typeSingle,
			value: keyword,
		})
	}

	return p.head, nil
}

func (p *patternParser) appendNode(n *node) {
	if p.head == nil {
		p.head = n
		p.tail = n
	} else {
		p.tail.next = n
		p.tail = n
	}
}

type patternCompiler struct {
	node         *node
	converterMap map[string]NewConverter

	head Converter
	tail Converter
}

// NewPatternCompiler creates a new instance of pattern compiler.
func NewPatternCompiler(node *node, converterMap map[string]NewConverter) *patternCompiler {
	return &patternCompiler{
		node:         node,
		converterMap: converterMap,
	}
}

// Compile will compile pattern to converter.
func (p *patternCompiler) Compile() (Converter, error) {
	for n := p.node; n != nil; n = n.next {
		switch n._type {
		case typeLiteral:
			p.appendConverter(NewLiteralConverter(n.value))

		case typeSingle:
			newConverter, ok := p.converterMap[n.value]
			if ok {
				c := newConverter()
				c.AttachOptions(n.options)
				p.appendConverter(c)
			} else {
				return nil, fmt.Errorf("failed to resolve converter for [%v]", n.value)
			}

		case typeComposite:
			newConverter, ok := p.converterMap[n.value]
			if ok {
				compositeConverter := newConverter()
				compositeConverter.AttachOptions(n.options)
				childCompiler := NewPatternCompiler(n.child, p.converterMap)
				childConverter, err := childCompiler.Compile()
				if err != nil {
					return nil, err
				}
				compositeConverter.AttachChild(childConverter)
				p.appendConverter(compositeConverter)
			} else {
				return nil, fmt.Errorf("failed to resolve converter for [%v]", n.value)
			}
		}
	}

	return p.head, nil
}

func (p *patternCompiler) appendConverter(c Converter) {
	if p.head == nil {
		p.head = c
		p.tail = c
	} else {
		p.tail.AttatchNext(c)
		p.tail = c
	}
}

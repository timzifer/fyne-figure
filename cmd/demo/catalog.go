package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/timzifer/fyne_figure/cmd/demo/plots"
)

// entry is one leaf of the tree: a chart, and how to build it on stage.
type entry struct {
	id, group, title, note string
	build                  func(env) (*view, error)
}

// catalogue is the tree: groups in order, and the entries under each.
type catalogue struct {
	entries  []*entry
	byID     map[string]*entry
	children map[string][]string // tree node → child nodes; "" is the root
	groups   map[string]string   // group node → group name
}

// groupPrefix marks a tree node as a group rather than an entry, so that the
// two can never be confused whatever an entry is called.
const groupPrefix = "group:"

// catalog is every chart figure draws — package plots — and the ones that
// need more than one widget or a clock — showcase.go — in the tree's order.
func catalog() *catalogue {
	var all []*entry
	for _, pe := range plots.All() {
		all = append(all, fromPlots(pe))
	}
	all = append(all, showcase()...)

	c := &catalogue{
		byID:     map[string]*entry{},
		children: map[string][]string{},
		groups:   map[string]string{},
	}
	for _, g := range append(plots.Groups(), groupInteraction) {
		node := groupPrefix + g
		c.groups[node] = g
		for _, e := range all {
			if e.group != g {
				continue
			}
			if _, dup := c.byID[e.id]; dup {
				log.Printf("catalogue: %q is listed twice; showing the first", e.id)
				continue
			}
			c.byID[e.id] = e
			c.entries = append(c.entries, e)
			c.children[node] = append(c.children[node], e.id)
		}
		if len(c.children[node]) > 0 {
			c.children[""] = append(c.children[""], node)
		}
	}
	for _, e := range all {
		if c.byID[e.id] == nil {
			log.Printf("catalogue: %q is in group %q, which is not in the tree", e.id, e.group)
		}
	}
	return c
}

// fromPlots makes a package plots entry buildable: a flat chart goes in a
// chart widget, a scene in an orbit widget and a grid in a picture.
func fromPlots(pe plots.Entry) *entry {
	e := &entry{id: pe.ID, group: pe.Group, title: pe.Title, note: pe.Note}
	switch {
	case pe.Plot != nil:
		e.build = func(en env) (*view, error) { return flatView(en, pe.Plot()), nil }
	case pe.Scene != nil:
		e.build = func(en env) (*view, error) { return orbitView(en, pe.Scene()), nil }
	case pe.Grid != nil:
		e.build = func(env) (*view, error) { return gridView(pe.Grid()), nil }
	default:
		e.build = func(env) (*view, error) { return nil, errors.New("the entry has nothing to build") }
	}
	return e
}

// find is the entry with the given id, or the first one.
func (c *catalogue) find(id string) *entry {
	if e, ok := c.byID[id]; ok {
		return e
	}
	return c.entries[0]
}

func (c *catalogue) isBranch(id string) bool {
	return id == "" || strings.HasPrefix(id, groupPrefix)
}

func (c *catalogue) label(id string) string {
	if g, ok := c.groups[id]; ok {
		return fmt.Sprintf("%s  (%d)", g, len(c.children[id]))
	}
	if e, ok := c.byID[id]; ok {
		return e.title
	}
	return id
}

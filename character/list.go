package character

import (
	"container/list"
	"iter"
	"slices"

	"github.com/dqso/ludum-dare-59/entity"
)

type List struct {
	list *list.List
}

func NewList() *List {
	return &List{
		list: list.New(),
	}
}

func (l *List) Add(c entity.Character) {
	l.list.PushBack(c)
}

func (l *List) All() iter.Seq[entity.Character] {
	return func(yield func(entity.Character) bool) {
		for e := l.list.Front(); e != nil; e = e.Next() {
			if !yield(e.Value.(entity.Character)) {
				return
			}
		}
	}
}

func (l *List) FilterByRoles(roles ...entity.CharacterRole) iter.Seq[entity.Character] {
	return func(yield func(entity.Character) bool) {
		for e := l.list.Front(); e != nil; e = e.Next() {
			value := e.Value.(entity.Character)
			if slices.Contains(roles, value.Role()) {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (l *List) FilterByCompany(company string) iter.Seq[entity.Character] {
	return func(yield func(entity.Character) bool) {
		for e := l.list.Front(); e != nil; e = e.Next() {
			value := e.Value.(entity.Character)
			if value.Company() == company {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (l *List) FilterFunc(fn func(c entity.Character) bool) iter.Seq[entity.Character] {
	return func(yield func(entity.Character) bool) {
		for e := l.list.Front(); e != nil; e = e.Next() {
			value := e.Value.(entity.Character)
			if fn(value) {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func FilterIdle(c entity.Character) bool {
	return c.GetQuestion() == nil && c.InterviewResult() == nil
}

func (l *List) DeleteFunc(fn func(c entity.Character) bool) iter.Seq2[entity.Character, bool] {
	return func(yield func(entity.Character, bool) bool) {
		for e := l.list.Front(); e != nil; {
			next := e.Next()
			c := e.Value.(entity.Character)
			deleted := fn(c)
			if deleted {
				l.list.Remove(e)
			}
			if !yield(c, deleted) {
				return
			}
			e = next
		}
	}
}

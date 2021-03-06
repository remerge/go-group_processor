package groupprocessor

import (
	"fmt"

	"github.com/remerge/cue"
)

type LoadSaver interface {
	Load(msg interface{}) Processable
	Save(p Processable) error
	Done(p Processable) bool
	Fail(p Processable, err error) bool
}

type DefaultLoadSaver struct {
	Name string
	Log  cue.Logger
}

func (ls *DefaultLoadSaver) New(name string) error {
	ls.Name = name
	ls.Log = cue.NewLogger(ls.Name)
	return nil
}

func (ls *DefaultLoadSaver) Load(value interface{}) Processable {
	return &DefaultProcessable{
		value: value,
	}
}

func (ls *DefaultLoadSaver) Save(_ Processable) error {
	return nil
}

func (ls *DefaultLoadSaver) Done(_ Processable) bool {
	return true
}

func (ls *DefaultLoadSaver) Fail(p Processable, err error) bool {
	var value interface{}
	if p != nil {
		value = p.Value()
	} else {
		value = "<nil>"
	}
	if ls == nil {
		fmt.Printf("DefaultLoadSaver is nil, value: %#v", value)
		return false
	}
	if ls.Log == nil {
		fmt.Printf("DefaultLoadSaver logger is nil, value: %#v", value)
		return false
	}
	// nolint: errcheck
	ls.Log.WithFields(cue.Fields{
		"value": value,
	}).Error(err, "failed to process message")

	return false
}

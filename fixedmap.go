package main

import (
    "container/list"
)

type FixedMap struct {
    Map  map[string]*MapElement
    List *list.List
    Size int
}

type MapElement struct {
    Element *list.Element
    Value   *File
}

func NewFixedMap(size int) *FixedMap {
    fm := new(FixedMap)
    fm.Map = make(map[string]*MapElement)
    fm.List = list.New()
    return fm
}

func (fm *FixedMap) Get(key string) *File {
    elem, ok := fm.Map[key]
    if ok {
        return elem.Value
    } else {
        return nil
    }
}

func (fm *FixedMap) Put(key string, value *File) {
    element := fm.List.PushFront(key)
    fm.Map[key] = &MapElement{ element, value }

    if fm.List.Len() == fm.Size {
        element = fm.List.Back()
        key, _ := element.Value.(string)
        delete(fm.Map, key)
        fm.List.Remove(element)
    }
}

func (fm *FixedMap) Remove(key string) {
    elem, ok := fm.Map[key]
    if !ok {
        return
    }

    delete(fm.Map, key)
    fm.List.Remove(elem.Element)
}

package main

import (
    "container/list"
)

/* FixedMap:
 * A fixed size map that pushes the last
 * used value from the stack if size limit
 * is reached and user attempts .Put()
 */
type FixedMap struct {
    Map  map[string]*MapElement
    List *list.List
    Size int
}

/* MapElement:
 * Simple structure to wrap pointer to list
 * element and stored map value together.
 */
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
        /* We're at capacity! SIR! */
        element = fm.List.Back()

        /* We don't check here as we know this is ALWAYS a string */
        key, _ := element.Value.(string)

        /* Finally delete the map entry and list element! */
        delete(fm.Map, key)
        fm.List.Remove(element)
    }
}

func (fm *FixedMap) Remove(key string) {
    elem, ok := fm.Map[key]
    if !ok {
        /* We don't have this key, return */
        return
    }

    /* Remove the selected element */
    delete(fm.Map, key)
    fm.List.Remove(elem.Element)
}

package main

import (
    "container/list"
)

/* TODO: work on efficiency. use our own lower level data structure? */

/* FixedMap:
 * A fixed size map that pushes the last
 * used value from the stack if size limit
 * is reached.
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
    return &FixedMap{
        make(map[string]*MapElement),
        list.New(),
        size,
    }
}

/* Get file in map for key, or nil */
func (fm *FixedMap) Get(key string) *File {
    elem, ok := fm.Map[key]
    if ok {
        return elem.Value
    } else {
        return nil
    }
}

/* Put file in map as key, pushing out last file
 * if size limit reached */
func (fm *FixedMap) Put(key string, value *File) {
    element := fm.List.PushFront(key)
    fm.Map[key] = &MapElement{ element, value }

    if fm.List.Len() > fm.Size {
        /* We're at capacity! SIR! */
        element = fm.List.Back()

        /* We don't check here as we know this is ALWAYS a string */
        key, _ := element.Value.(string)

        /* Finally delete the map entry and list element! */
        delete(fm.Map, key)
        fm.List.Remove(element)

        Config.LogSystem("Popped key: %s\n", key)
    }
}

/* Try delete element, else do nothing */
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

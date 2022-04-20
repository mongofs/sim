/*
 * Copyright 2022 steven
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *    http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sortmap

import (
	"errors"
	"sync"
)

type sortMap struct {
	rw  sync.RWMutex
	lastIndex int
	sort []interface{}
	search map[string]*node

}

type node struct {
	index int
	data interface{}
}


func Map() *sortMap {
	return &sortMap{
		rw:     sync.RWMutex{},
		sort:   nil,
		search: map[string]*node{},
	}
}


func (s *sortMap) Add (tag string, data interface{}) (string,error){
	s.rw.Lock()
	defer s.rw.Unlock()
	if _, ok := s.search[tag] ;ok {
		return "",errors.New( "sortmap : tag is already exist")
	}
	s.sort = append(s.sort, &node{})
	return "", nil
}


func (s *sortMap) Del (tag string) {


}


func (s *sortMap) List (limit , page int){


}
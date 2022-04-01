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

package sim

import (
	"net/http"
)

type Cli struct {
	Connect
	reader        *http.Request
}


func NewClient(w http.ResponseWriter, r *http.Request, closeSig chan<- string, token *string, option *Option) (Client, error) {
	res := &Cli{
		reader:        r,
	}

	conn, err := NewGorilla(token, closeSig, option, w, r,option.ServerReceive)
	if err != nil {
		return nil, err
	}
	res.Connect = conn
	return res, nil
}

func (c *Cli) Request() *http.Request {
	return c.reader
}

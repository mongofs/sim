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
	"errors"
	"fmt"
)

type  DefaultValidate struct {}


func (d *DefaultValidate) Validate(token string)error{
	if token == "" {
		return errors.New("token is not good ")
	}
	return nil
}


func (d *DefaultValidate)ValidateFailed(err error,cli Client){

	fmt.Println(err.Error())
	// 当用户登录验证失败，逻辑应该在这里来处理
	cli.Send([]byte("user validate is bad"))
	cli.Offline()
}


func (d *DefaultValidate)ValidateSuccess(cli Client){
	// 当用户登录验证失败，逻辑应该在这里来处理
	cli.Send([]byte("user is online "))
}


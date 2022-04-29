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

package errors

import (
	"errors"
)


var (
	ErrUserExist =errors.New("im/bucket : Cannot login repeatedly")
	ErrCliISNil  =errors.New("im/bucket : client is not online ")
)


var (
	ErrTokenIsNil = errors.New("sim : ValidateKey can't be nil")
	ErrUserBufferIsFull = errors.New("sim : The client buffer is about to fill up")


	ErrBadParam = errors.New("sim :  bad param ")


	ERRWTINotStartServer = errors.New("sim/label: not start label example ")
	ERRWTIGroupNotClear = errors.New("sim/label : group is not clear ")
	ERRWTITargetNotExist = errors.New("sim/label : label is not exist")
)
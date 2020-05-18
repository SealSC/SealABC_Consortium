/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package enum

import (
    "reflect"
    "strconv"
)

const (
    errorCodeTag    = "code"
    errorMessageTag = "msg"
)

type ErrorElement struct {
    code    int64
    name    string
    message string
    data    interface{}
}

type ErrorElementInstance struct {
    ErrorElement
}

func (e *ErrorElement)Instance() *ErrorElementInstance {
    return e.InstanceWithData("", nil)
}

func (e *ErrorElement) InstanceWithData(msg string, data interface{}) *ErrorElementInstance {
    return &ErrorElementInstance{
            ErrorElement{
                code:    e.code,
                name:    e.name,
                message: msg,
                data:    data,
            },
    }
}

func (e *ErrorElement)Code() int64  {
    return e.code
}

func (e *ErrorElement)Name() string  {
    return e.name
}

func (e *ErrorElement)Message() string  {
    return e.message
}

func (e *ErrorElement)Data() interface{}  {
    return e.data
}

func (e *ErrorElementInstance)Error() string  {
    return e.message
}

func BuildErrorEnum(enum interface{}) {
    buildEnum(enum, func(code int, name string, tag reflect.StructTag) reflect.Value {
        codeStr := tag.Get(errorCodeTag)
        codeNum := int64(code)
        if "" != codeStr {
            codeNum, _ = strconv.ParseInt(codeStr, 0, 64)
        }

        return reflect.ValueOf(ErrorElement{
            code: codeNum,
            name: name,
            message: tag.Get(errorMessageTag),
        })
    })
    return
}

// Copyright (c) 2017 OysterPack, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oysterpack_cloud_go

import (
	"testing"
	"fmt"
)

type Foo int

type Bar int

func TestTypeConversion(t *testing.T) {
	var foo Foo = 1
	var bar Bar = 2
	var fooBar Foo = Foo(bar)
	fmt.Println(foo,bar,fooBar)
}

package json

import (
	gojson "encoding/json"
	"reflect"
	"testing"

	"github.com/yuin/gopher-lua"
)

func TestSimple(t *testing.T) {
	const str = `
	local json = require("json")
	assert(type(json) == "table")
	assert(type(json.decode) == "function")
	assert(type(json.encode) == "function")

	assert(json.encode(true) == "true")
	assert(json.encode(1) == "1")
	assert(json.encode(-10) == "-10")
	assert(json.encode(nil) == "{}")

	local obj = {"a",1,"b",2,"c",3}
	local jsonStr = json.encode(obj)
	local jsonObj = json.decode(jsonStr)
	for i = 1, #obj do
		assert(obj[i] == jsonObj[i])
	end

	local obj = {name="Tim",number=12345}
	local jsonStr = json.encode(obj)
	local jsonObj = json.decode(jsonStr)
	assert(obj.name == jsonObj.name)
	assert(obj.number == jsonObj.number)

	local obj = {"a","b",what="c",[5]="asd"}
	local jsonStr = json.encode(obj)
	local jsonObj = json.decode(jsonStr)
	assert(obj[1] == jsonObj["1"])
	assert(obj[2] == jsonObj["2"])
	assert(obj.what == jsonObj["what"])
	assert(obj[5] == jsonObj["5"])

	assert(json.decode("null") == nil)

	assert(json.decode(json.encode({person={name = "tim",}})).person.name == "tim")

	local obj = {
		abc = 123,
		def = nil,
	}
	local obj2 = {
		obj = obj,
	}
	obj.obj2 = obj2
	assert(json.encode(obj) == nil)

	local a = {}
	for i=1, 5 do
		a[i] = i
	end
	assert(json.encode(a) == "[1,2,3,4,5]")
	`
	s := lua.NewState()
	Preload(s)
	if err := s.DoString(str); err != nil {
		t.Error(err)
	}
}

func TestEmptyMapArray(t *testing.T) {
	const code = `
local json = require("JSON")

function test(input)
	obj = json.decode(input)
	return json.encode(obj)
end
`

	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("JSON", Loader)
	if err := L.DoString(code); err != nil {
		t.Error(err)
	}

	type genericObject map[string]interface{}

	obj := genericObject{
		"emptyObject": genericObject{},
		"emptyArray": []string{},
	}

	raw, err := gojson.Marshal(obj)
	if err != nil {
		t.Error(err)
	}
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test"),
		NRet:    1,
		Protect: true,
	}, lua.LString(string(raw))); err != nil {
		t.Error(err)
	}
	ret := L.Get(-1) // returned value
	L.Pop(1)          // remove received value

	var obj2 genericObject
	err = gojson.Unmarshal([]byte(ret.String()), &obj2)
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(obj, obj2) == false {
		t.Errorf("Objects differ: in=%s out=%s", raw, ret.String())
	}

}

func TestCustomRequire(t *testing.T) {
	const str = `
	local j = require("JSON")
	assert(type(j) == "table")
	assert(type(j.decode) == "function")
	assert(type(j.encode) == "function")
	`
	s := lua.NewState()
	s.PreloadModule("JSON", Loader)
	if err := s.DoString(str); err != nil {
		t.Error(err)
	}
}

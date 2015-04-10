package json

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/yuin/gopher-lua"
)

var (
	errFunction = errors.New("cannot convert function to string")
	errChannel  = errors.New("cannot convert channel to string")
	errUserData = errors.New("cannot convert userdata to string")
)

type jsonValue struct {
	lua.LValue
}

func (j jsonValue) MarshalJSON() ([]byte, error) {
	return toJSON(j.LValue)
}

func toJSON(value lua.LValue) (data []byte, err error) {
	switch converted := value.(type) {
	case lua.LBool:
		data, err = json.Marshal(converted)
	case lua.LChannel:
		err = errChannel
	case lua.LNumber:
		data, err = json.Marshal(converted)
	case *lua.LFunction:
		err = errFunction
	case *lua.LNilType:
		data, err = json.Marshal(converted)
	case lua.LString:
		data, err = json.Marshal(converted)
	case *lua.LTable:
		var arr []jsonValue
		var obj map[string]jsonValue

		converted.ForEach(func(k lua.LValue, v lua.LValue) {
			i, numberKey := k.(lua.LNumber)
			if numberKey && obj == nil {
				index := int(i) - 1
				if index != len(arr) {
					// map out of order; convert to map
					obj = make(map[string]jsonValue)
					for i, value := range arr {
						obj[strconv.Itoa(i+1)] = value
					}
					obj[strconv.Itoa(index+1)] = jsonValue{v}
					return
				}
				arr = append(arr, jsonValue{v})
				return
			}
			if obj == nil {
				obj = make(map[string]jsonValue)
				for i, value := range arr {
					obj[strconv.Itoa(i+1)] = value
				}
			}
			obj[k.String()] = jsonValue{v}
		})
		if obj != nil {
			data, err = json.Marshal(obj)
		} else {
			data, err = json.Marshal(arr)
		}
	case *lua.LUserData:
		// TODO: call metatable __tostring?
		err = errUserData
	}
	return
}

func fromJSON(L *lua.LState, value interface{}) lua.LValue {
	switch converted := value.(type) {
	case bool:
		return lua.LBool(converted)
	case float64:
		return lua.LNumber(converted)
	case string:
		return lua.LString(converted)
	case []interface{}:
		arr := L.NewTable()
		for _, item := range converted {
			arr.Append(fromJSON(L, item))
		}
		return arr
	case map[string]interface{}:
		tbl := L.NewTable()
		for key, item := range converted {
			tbl.RawSetH(lua.LString(key), fromJSON(L, item))
		}
		return tbl
	}
	return lua.LNil
}

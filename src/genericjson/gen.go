package genericjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
)

type Any interface{}
type GenJson struct{ Any }

var NotExists = errors.New("incorrect or not existent json path")

func FromGeneric(any interface{}) GenJson { return GenJson{Any: any} }

func (self *GenJson) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &self.Any)
}

func (self GenJson) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.Any)
}

func (self GenJson) Bool(args ...interface{}) (retval bool, err error) {
	s, err := self.Unwind(args...)
	if err == nil {
		var ok bool
		retval, ok = s.Any.(bool)
		debug(fmt.Sprintf("Bool: s is '%v', retval is %v", s.Any, retval))
		if !ok {
			return false, errors.New(fmt.Sprintf("value %s '%v' is not a bool!", args, s.Any))
		}
	}
	return
}

func (self GenJson) Int(args ...interface{}) (retval int, err error) {
	s, err := self.Unwind(args...)
	if err == nil {
		r, ok := s.Any.(float64)
		debug(fmt.Sprintf("Int: s is '%v', retval is %v, %T", s.Any, r, s.Any))
		if !ok {
			return 0, errors.New(fmt.Sprintf("value %s '%v' is not an int!", args, s.Any))
		}
		if math.Abs(r-math.Trunc(r)) > math.SmallestNonzeroFloat64 {
			return 0, errors.New(fmt.Sprintf("value '%v' is a float and not an int!", s.Any))
		}
		retval = int(r)
	}
	return
}

func (self GenJson) Float(args ...interface{}) (retval float64, err error) {
	s, err := self.Unwind(args...)
	if err == nil {
		var ok bool
		retval, ok = s.Any.(float64)
		debug(fmt.Sprintf("Float: s is '%v', retval is %v", s.Any, retval))
		if !ok {
			return math.NaN(), errors.New(fmt.Sprintf("value %s '%v' is not an float64!", args, s.Any))
		}
	}
	return
}

func (self GenJson) Empty() bool {
	return self.Any == nil
}

func (self GenJson) String(args ...interface{}) (retval string, err error) {
	s, err := self.Unwind(args...)
	if err == nil {
		var ok bool
		retval, ok = s.Any.(string)
		debug(fmt.Sprintf("String: s is '%v', retval is %s", s.Any, retval))
		if !ok {
			return "", errors.New(fmt.Sprintf("value %s '%v' is not a string!", args, s.Any))
		}
	}
	return
}

func (self GenJson) StringOrEmpty(args ...interface{}) string {
	retval, err := self.String(args...)
	if err != nil {
		return ""
	}
	return retval
}

func (self GenJson) UnwindOrNil(args ...interface{}) GenJson {
	retval, err := self.Unwind(args...)
	if err != nil {
		return FromGeneric(nil)
	}
	return retval
}

func (self GenJson) Array(args ...interface{}) (retval []interface{}, err error) {
	s, err := self.Unwind(args...)
	if err == nil {
		var ok bool
		retval, ok = s.Any.([]interface{})
		debug(fmt.Sprintf("String: s is '%v', retval is %s", s.Any, retval))
		if !ok {
			return []interface{}{}, errors.New(fmt.Sprintf("value %s '%v' is not an array!", args, s.Any))
		}
	}
	return
}

func (self GenJson) ArrayOrEmpty(args ...interface{}) []interface{} {
	if a, e := self.Array(args...); e == nil {
		return a
	}
	return []interface{}{}
}

func (self GenJson) Unwind(args ...interface{}) (s GenJson, err error) {
	s = self
	for _, arg := range args {
		switch v := s.Any.(type) {
		case []interface{}:
			i, ok := arg.(int)
			if !ok {
				return s, errors.New(fmt.Sprintf("expected integer, found `%v`", arg))
			}
			if i < 0 || i >= len(v) {
				return s, errors.New(fmt.Sprintf("index out of bounds %d", i))
			}
			s.Any = v[i]
		case map[string]interface{}:
			str, ok := arg.(string)
			if !ok {
				return s, errors.New(fmt.Sprintf("expected string, found `%v`", arg))
			}
			s.Any = v[str]
		default:
			return s, errors.New(fmt.Sprintf("incorrect or not existent json path '%v'", args))
		}
	}
	return
}

func (self GenJson) Delete(args ...interface{}) (err error) {
	if len(args) == 0 {
		return NotExists
	}
	s := self
	if len(args) > 1 {
		s, err = self.Unwind(args[0 : len(args)-1]...)
		if err != nil {
			return
		}
	}
	arg := args[len(args)-1]
	switch v := s.Any.(type) {
	case []interface{}:
		i, ok := arg.(int)
		if !ok {
			return errors.New(fmt.Sprintf("expected integer, found `%v`", arg))
		}
		if i < 0 || i >= len(v) {
			return errors.New(fmt.Sprintf("index out of bounds %d", i))
		}
		vn := make([]interface{}, len(v)-1)
		for k, idx := 0, 0; k < len(v); k++ {
			if i != k {
				vn[idx] = v[k]
				idx++
			}
		}
		self.Set(append([]interface{}{vn}, args[0:len(args)-1]...)...)
	case map[string]interface{}:
		str, ok := arg.(string)
		if !ok {
			return errors.New(fmt.Sprintf("expected string, found `%v`", arg))
		}
		delete(v, str)
	default:
		return NotExists
	}
	return nil
}

func (self GenJson) Clone() (s GenJson) {
	b, _ := self.MarshalJSON()
	s.UnmarshalJSON(b)
	return
}

func (self GenJson) ScanObject(predicate func(interface{}) bool,
	args ...interface{}) (GenJson, []interface{}, bool) {
	path, s := []interface{}{}, self
	for idx, arg := range args {
		if len(args)-1 == idx {
			if predicate(s.Any) {
				return s, path, true
			}
		}
		path = append(path, arg)
		//log.Printf("ScanObject %v %v", arg, s )
		switch v := s.Any.(type) {
		case []interface{}:
			i, ok := arg.(int)
			if !ok {
				return s, path, false
			}
			if i == -1 {
				if len(args)-1 == idx {
					for j, obj := range v {
						s.Any = obj
						if predicate(obj) {
							path[len(path)-1] = j
							return s, path, true
						}
					}
				} else {
					rest := args[(idx + 1):]
					for j, obj := range v {
						s.Any = obj
						retval, par, ok := s.ScanObject(predicate, rest...)
						if ok {
							path[len(path)-1] = j
							path = append(path, par...)
							return retval, path, ok
						}
					}
				}
			} else if i < 0 || i >= len(v) {
				return s, path, false
			} else {
				s1, p1, ok := s.ScanObject(predicate, args[(idx+1):]...)
				if ok {
					path := append(path, p1...)
					return s1, path, ok
				}
			}
		case map[string]interface{}:
			str, ok := arg.(string)
			if !ok {
				return s, path, false
			}
			s.Any = v[str]
		default:
			return s, path, false
		}
	}
	return s, path, false
}

func (self GenJson) Set(args ...interface{}) error {
	s, err := self.Unwind(args[1:(len(args) - 1)]...)
	val := args[0]
	if err == nil {
		arg := args[len(args)-1]
		switch v := s.Any.(type) {
		case []interface{}:
			i, ok := arg.(int)
			if !ok {
				return errors.New(fmt.Sprintf("expected integer, found `%v`", arg))
			}
			if i < 0 || i >= len(v) {
				return errors.New(fmt.Sprintf("index out of bounds %d", i))
			}
			v[i] = val
			return nil
		case map[string]interface{}:
			str, ok := arg.(string)
			if !ok {
				return errors.New(fmt.Sprintf("expected string, found `%v`", arg))
			}
			v[str] = val
			return nil
		default:
			return NotExists
		}
	}
	return err
}

var d = false

func SetDebug(b bool) { d = b }
func debug(message string) {
	if d {
		log.Println(message)
	}
}

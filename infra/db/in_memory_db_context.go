package db

import (
	"encoding/json"
	"errors"
	"reflect"
)

type inMemorySnapshots map[string][]byte

type InMemoryDBContext struct {
	HelloWorldKV *HelloWorldKv `repository:"true"`
}

func NewInMemoryDBContext(
	helloWorldKV *HelloWorldKv,
) *InMemoryDBContext {
	return &InMemoryDBContext{
		HelloWorldKV: helloWorldKV,
	}
}

func (db *InMemoryDBContext) CreateSnap() ([]byte, error) {
	v := reflect.ValueOf(db).Elem()
	var snapshots inMemorySnapshots = make(inMemorySnapshots)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)
		if fieldType.Tag.Get("repository") != "true" {
			continue
		}
		if field.Kind() != reflect.Ptr || field.Elem().Kind() != reflect.Struct {
			continue
		}

		var name string
		methodGetName := field.MethodByName("GetName")
		if methodGetName.IsValid() &&
			methodGetName.Type().NumIn() == 0 &&
			methodGetName.Type().NumOut() == 1 &&
			methodGetName.Type().Out(0).Kind() == reflect.String {
			name = methodGetName.Call(nil)[0].String()
		}
		if len(name) == 0 {
			return []byte{}, errors.New("create snap shot error check repository has name")
		}

		createSnapMethod := field.MethodByName("CreateSnap")
		if createSnapMethod.IsValid() &&
			createSnapMethod.Type().NumIn() == 0 &&
			createSnapMethod.Type().NumOut() == 2 &&
			createSnapMethod.Type().Out(0).Kind() == reflect.Slice &&
			createSnapMethod.Type().Out(0).Elem().Kind() == reflect.Uint8 &&
			createSnapMethod.Type().Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			results := createSnapMethod.Call(nil)
			if err, ok := results[1].Interface().(error); ok && err != nil {
				return nil, err
			}
			bz := results[0].Interface().([]byte)
			if len(bz) == 0 {
				continue
			}

			snapshots[name] = results[0].Interface().([]byte)
		}
	}

	if len(snapshots) == 0 {
		return []byte{}, nil
	}

	bz, err := json.Marshal(snapshots)
	if err != nil {
		return []byte{}, err
	}
	return bz, nil
}

func (db *InMemoryDBContext) LoadSnap(bz []byte) error {
	var snapshots inMemorySnapshots = make(inMemorySnapshots)

	if err := json.Unmarshal(bz, &snapshots); err != nil {
		return err
	}

	if len(snapshots) == 0 {
		return nil
	}

	v := reflect.ValueOf(db).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		if fieldType.Tag.Get("repository") != "true" {
			continue
		}

		if field.Kind() != reflect.Ptr || field.Elem().Kind() != reflect.Struct {
			continue
		}
		var name string
		methodGetName := field.MethodByName("GetName")
		if methodGetName.IsValid() &&
			methodGetName.Type().NumIn() == 0 &&
			methodGetName.Type().NumOut() == 1 &&
			methodGetName.Type().Out(0).Kind() == reflect.String {
			name = methodGetName.Call(nil)[0].String()
		}
		if len(name) == 0 {
			return errors.New("create snap shot error check repository has name")
		}

		bz, existed := snapshots[name]
		if !existed {
			continue
		}

		methodLoadSnapshot := field.MethodByName("LoadSnap")
		if methodLoadSnapshot.IsValid() &&
			methodLoadSnapshot.Type().NumIn() == 1 &&
			methodLoadSnapshot.Type().In(0).Kind() == reflect.Slice &&
			methodLoadSnapshot.Type().In(0).Elem().Kind() == reflect.Uint8 &&
			methodLoadSnapshot.Type().NumOut() == 1 &&
			methodLoadSnapshot.Type().Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			results := methodLoadSnapshot.Call([]reflect.Value{reflect.ValueOf(bz)})
			if err, ok := results[0].Interface().(error); ok && err != nil {
				return err
			}
		}
	}

	return nil
}

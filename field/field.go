/*
 Copyright 2020 The Qmgo Authors.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package field

import (
	"context"
	"reflect"
	"time"

	"github.com/nguyenngodinh/qmgo/operator"
)

var nilTime time.Time

// filedHandler defines the relations between field type and handler
var fieldHandler = map[operator.OpType]func(ctx context.Context, doc interface{}) error{
	operator.BeforeInsert:  beforeInsert,
	operator.BeforeUpdate:  beforeUpdate,
	operator.BeforeReplace: beforeUpdate,
	operator.BeforeUpsert:  beforeUpsert,
}

//func init() {
//	middleware.Register(Do)
//}

// Do call the specific method to handle field based on fType
// Don't use opts here
func Do(ctx context.Context, doc interface{}, opType operator.OpType, opts ...interface{}) error {
	to := reflect.TypeOf(doc)
	if to == nil {
		return nil
	}
	switch reflect.TypeOf(doc).Kind() {
	case reflect.Slice:
		return sliceHandle(ctx, doc, opType)
	case reflect.Ptr:
		v := reflect.ValueOf(doc).Elem()
		switch v.Kind() {
		case reflect.Slice:
			return sliceHandle(ctx, v.Interface(), opType)
		default:
			return do(ctx, doc, opType)
		}
	}
	//fmt.Println("not support type")
	return nil
}

// sliceHandle handles the slice docs
func sliceHandle(ctx context.Context, docs interface{}, opType operator.OpType) error {
	// []interface{}{UserType{}...}
	if h, ok := docs.([]interface{}); ok {
		for _, v := range h {
			if err := do(ctx, v, opType); err != nil {
				return err
			}
		}
		return nil
	}
	// []UserType{}
	s := reflect.ValueOf(docs)
	for i := 0; i < s.Len(); i++ {
		if err := do(ctx, s.Index(i).Interface(), opType); err != nil {
			return err
		}
	}
	return nil
}

// beforeInsert handles field before insert
// If value of field createAt is valid in doc, upsert doesn't change it
// If value of field id is valid in doc, upsert doesn't change it
// Change the value of field updateAt anyway
func beforeInsert(ctx context.Context, doc interface{}) error {
	if ih, ok := doc.(DefaultFieldHook); ok {
		ih.DefaultId()
		ih.DefaultCreateAt()
		ih.DefaultUpdateAt()
	}
	if ih, ok := doc.(CustomFieldsHook); ok {
		fields := ih.CustomFields()
		fields.(*CustomFields).CustomId(ctx, doc)
		fields.(*CustomFields).CustomCreateTime(ctx, doc)
		fields.(*CustomFields).CustomCreateBy(ctx, doc)
		fields.(*CustomFields).CustomUpdateTime(ctx, doc)
		fields.(*CustomFields).CustomUpdateBy(ctx, doc)
	}
	return nil
}

// beforeUpdate handles field before update
func beforeUpdate(ctx context.Context, doc interface{}) error {
	if ih, ok := doc.(DefaultFieldHook); ok {
		ih.DefaultUpdateAt()
	}
	if ih, ok := doc.(CustomFieldsHook); ok {
		fields := ih.CustomFields()
		fields.(*CustomFields).CustomUpdateTime(ctx, doc)
		fields.(*CustomFields).CustomUpdateBy(ctx, doc)
	}
	return nil
}

// beforeUpsert handles field before upsert
// If value of field createAt is valid in doc, upsert doesn't change it
// If value of field id is valid in doc, upsert doesn't change it
// Change the value of field updateAt anyway
func beforeUpsert(ctx context.Context, doc interface{}) error {
	if ih, ok := doc.(DefaultFieldHook); ok {
		ih.DefaultId()
		ih.DefaultCreateAt()
		ih.DefaultUpdateAt()
	}
	if ih, ok := doc.(CustomFieldsHook); ok {
		fields := ih.CustomFields()
		fields.(*CustomFields).CustomId(ctx, doc)
		fields.(*CustomFields).CustomCreateTime(ctx, doc)
		fields.(*CustomFields).CustomCreateBy(ctx, doc)
		fields.(*CustomFields).CustomUpdateTime(ctx, doc)
		fields.(*CustomFields).CustomUpdateBy(ctx, doc)
	}
	return nil
}

// do check if opType is supported and call fieldHandler
func do(ctx context.Context, doc interface{}, opType operator.OpType) error {
	if f, ok := fieldHandler[opType]; !ok {
		return nil
	} else {
		return f(ctx, doc)
	}
}

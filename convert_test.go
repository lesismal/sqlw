// Copyright 2022 lesismal. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlw

import (
	"reflect"
	"strings"
	"testing"
)

func Benchmark_sqlMappingKey2(b *testing.B) {
	type args struct {
		opTyp string
		query string
		typ   reflect.Type
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	keyFunc := func(opTyp, query string, typ reflect.Type) string {
		var b strings.Builder
		b.WriteString(opTyp)
		b.WriteString(query)
		b.WriteString(typ.String())
		return b.String()
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			keyFunc(tt.args.opTyp, tt.args.query, tt.args.typ)
		}
	}
}

func Benchmark_sqlMappingKey(b *testing.B) {
	type args struct {
		opTyp string
		query string
		typ   reflect.Type
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			sqlMappingKey(tt.args.opTyp, tt.args.query, tt.args.typ)
		}
	}
}

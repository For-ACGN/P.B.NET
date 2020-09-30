package vm

import (
	"fmt"
	"testing"
)

func TestSlices(t *testing.T) {
	tests := []Test{
		{Script: `[1++]`, RunError: fmt.Errorf("invalid operation")},
		{Script: `1++[0]`, RunError: fmt.Errorf("invalid operation")},

		{Script: `[]`, RunOutput: []interface{}{}},
		{Script: `[nil]`, RunOutput: []interface{}{nil}},
		{Script: `[true]`, RunOutput: []interface{}{true}},
		{Script: `["a"]`, RunOutput: []interface{}{"a"}},
		{Script: `[1]`, RunOutput: []interface{}{int64(1)}},
		{Script: `[1.1]`, RunOutput: []interface{}{1.1}},

		{Script: `a = []; a.b`, RunError: fmt.Errorf("type slice does not support member operation")},
		{Script: `a = []; a.b = 1`, RunError: fmt.Errorf("type slice does not support member operation")},

		{Script: `a = []`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{}}},
		{Script: `a = [nil]`, RunOutput: []interface{}{interface{}(nil)}, Output: map[string]interface{}{"a": []interface{}{interface{}(nil)}}},
		{Script: `a = [true]`, RunOutput: []interface{}{true}, Output: map[string]interface{}{"a": []interface{}{true}}},
		{Script: `a = [1]`, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []interface{}{int64(1)}}},
		{Script: `a = [1.1]`, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": []interface{}{1.1}}},
		{Script: `a = ["a"]`, RunOutput: []interface{}{"a"}, Output: map[string]interface{}{"a": []interface{}{"a"}}},

		{Script: `a = [[]]`, RunOutput: []interface{}{[]interface{}{}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{}}}},
		{Script: `a = [[nil]]`, RunOutput: []interface{}{[]interface{}{interface{}(nil)}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{interface{}(nil)}}}},
		{Script: `a = [[true]]`, RunOutput: []interface{}{[]interface{}{true}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{true}}}},
		{Script: `a = [[1]]`, RunOutput: []interface{}{[]interface{}{int64(1)}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1)}}}},
		{Script: `a = [[1.1]]`, RunOutput: []interface{}{[]interface{}{1.1}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{1.1}}}},
		{Script: `a = [["a"]]`, RunOutput: []interface{}{[]interface{}{"a"}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{"a"}}}},

		{Script: `a = []; a += nil`, RunOutput: []interface{}{nil}, Output: map[string]interface{}{"a": []interface{}{nil}}},
		{Script: `a = []; a += true`, RunOutput: []interface{}{true}, Output: map[string]interface{}{"a": []interface{}{true}}},
		{Script: `a = []; a += 1`, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []interface{}{int64(1)}}},
		{Script: `a = []; a += 1.1`, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": []interface{}{1.1}}},
		{Script: `a = []; a += "a"`, RunOutput: []interface{}{"a"}, Output: map[string]interface{}{"a": []interface{}{"a"}}},

		{Script: `a = []; a += []`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{}}},
		{Script: `a = []; a += [nil]`, RunOutput: []interface{}{nil}, Output: map[string]interface{}{"a": []interface{}{nil}}},
		{Script: `a = []; a += [true]`, RunOutput: []interface{}{true}, Output: map[string]interface{}{"a": []interface{}{true}}},
		{Script: `a = []; a += [1]`, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []interface{}{int64(1)}}},
		{Script: `a = []; a += [1.1]`, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": []interface{}{1.1}}},
		{Script: `a = []; a += ["a"]`, RunOutput: []interface{}{"a"}, Output: map[string]interface{}{"a": []interface{}{"a"}}},

		{Script: `a = [0]; a[0]++`, RunOutput: int64(1), Output: map[string]interface{}{"a": []interface{}{int64(1)}}},
		{Script: `a = [[0]]; a[0][0]++`, RunOutput: int64(1), Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1)}}}},

		{Script: `a = [2]; a[0]--`, RunOutput: int64(1), Output: map[string]interface{}{"a": []interface{}{int64(1)}}},
		{Script: `a = [[2]]; a[0][0]--`, RunOutput: int64(1), Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1)}}}},

		{Script: `a`, Input: map[string]interface{}{"a": []bool{}}, RunOutput: []bool{}, Output: map[string]interface{}{"a": []bool{}}},
		{Script: `a`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []int32{}, Output: map[string]interface{}{"a": []int32{}}},
		{Script: `a`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []int64{}, Output: map[string]interface{}{"a": []int64{}}},
		{Script: `a`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []float32{}, Output: map[string]interface{}{"a": []float32{}}},
		{Script: `a`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []float64{}, Output: map[string]interface{}{"a": []float64{}}},
		{Script: `a`, Input: map[string]interface{}{"a": []string{}}, RunOutput: []string{}, Output: map[string]interface{}{"a": []string{}}},

		{Script: `a`, Input: map[string]interface{}{"a": []bool{true, false}}, RunOutput: []bool{true, false}, Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: []int32{1, 2}, Output: map[string]interface{}{"a": []int32{1, 2}}},
		{Script: `a`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: []int64{1, 2}, Output: map[string]interface{}{"a": []int64{1, 2}}},
		{Script: `a`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: []float32{1.1, 2.2}, Output: map[string]interface{}{"a": []float32{1.1, 2.2}}},
		{Script: `a`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: []float64{1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}}},
		{Script: `a`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: []string{"a", "b"}, Output: map[string]interface{}{"a": []string{"a", "b"}}},

		{Script: `a += true`, Input: map[string]interface{}{"a": []bool{}}, RunOutput: []bool{true}, Output: map[string]interface{}{"a": []bool{true}}},
		{Script: `a += 1`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []int32{1}, Output: map[string]interface{}{"a": []int32{1}}},
		{Script: `a += 1.1`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []int32{1}, Output: map[string]interface{}{"a": []int32{1}}},
		{Script: `a += 1`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []int64{1}, Output: map[string]interface{}{"a": []int64{1}}},
		{Script: `a += 1.1`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []int64{1}, Output: map[string]interface{}{"a": []int64{1}}},
		{Script: `a += 1`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []float32{1}, Output: map[string]interface{}{"a": []float32{1}}},
		{Script: `a += 1.1`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []float32{1.1}, Output: map[string]interface{}{"a": []float32{1.1}}},
		{Script: `a += 1`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []float64{1}, Output: map[string]interface{}{"a": []float64{1}}},
		{Script: `a += 1.1`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []float64{1.1}, Output: map[string]interface{}{"a": []float64{1.1}}},
		{Script: `a += "a"`, Input: map[string]interface{}{"a": []string{}}, RunOutput: []string{"a"}, Output: map[string]interface{}{"a": []string{"a"}}},
		{Script: `a += 97`, Input: map[string]interface{}{"a": []string{}}, RunOutput: []string{"a"}, Output: map[string]interface{}{"a": []string{"a"}}},

		{Script: `a[0]`, Input: map[string]interface{}{"a": []bool{true, false}}, RunOutput: true, Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: int32(1), Output: map[string]interface{}{"a": []int32{1, 2}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: int64(1), Output: map[string]interface{}{"a": []int64{1, 2}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: float32(1.1), Output: map[string]interface{}{"a": []float32{1.1, 2.2}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: 1.1, Output: map[string]interface{}{"a": []float64{1.1, 2.2}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: "a", Output: map[string]interface{}{"a": []string{"a", "b"}}},

		{Script: `a[1]`, Input: map[string]interface{}{"a": []bool{true, false}}, RunOutput: false, Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: int32(2), Output: map[string]interface{}{"a": []int32{1, 2}}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: int64(2), Output: map[string]interface{}{"a": []int64{1, 2}}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: float32(2.2), Output: map[string]interface{}{"a": []float32{1.1, 2.2}}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: 2.2, Output: map[string]interface{}{"a": []float64{1.1, 2.2}}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: "b", Output: map[string]interface{}{"a": []string{"a", "b"}}},

		{Script: `a[0]`, Input: map[string]interface{}{"a": []bool{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []bool{}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []int32{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []int32{}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []int64{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []int64{}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []float32{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []float32{}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []float64{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []float64{}}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": []string{}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []string{}}},

		{Script: `a[1] = true`, Input: map[string]interface{}{"a": []bool{true, false}}, RunOutput: true, Output: map[string]interface{}{"a": []bool{true, true}}},
		{Script: `a[1] = 3`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: int64(3), Output: map[string]interface{}{"a": []int32{1, 3}}},
		{Script: `a[1] = 3`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: int64(3), Output: map[string]interface{}{"a": []int64{1, 3}}},
		{Script: `a[1] = 3.3`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: 3.3, Output: map[string]interface{}{"a": []float32{1.1, 3.3}}},
		{Script: `a[1] = 3.3`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: 3.3, Output: map[string]interface{}{"a": []float64{1.1, 3.3}}},
		{Script: `a[1] = "c"`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: "c", Output: map[string]interface{}{"a": []string{"a", "c"}}},

		{Script: `a = []; a[0]`, RunError: fmt.Errorf("index out of range")},
		{Script: `a = []; a[-1]`, RunError: fmt.Errorf("index out of range")},
		{Script: `a = []; a[1] = 1`, RunError: fmt.Errorf("index out of range")},
		{Script: `a = []; a[-1] = 1`, RunError: fmt.Errorf("index out of range")},

		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": nil}, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": true}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": 1}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": int32(1)}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": int64(1)}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": float32(1.1)}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": 1.1}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": "1"}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": "a"}, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},

		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": testVarBoolP}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": testVarInt32P}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": testVarInt64P}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": testVarFloat32P}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": testVarFloat64P}, RunOutput: int64(2), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a]`, Input: map[string]interface{}{"a": testVarStringP}, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},

		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": nil}, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": true}, RunOutput: int64(3), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": 1}, RunOutput: int64(3), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": int32(1)}, RunOutput: int64(3), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": int64(1)}, RunOutput: int64(3), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": float32(1.1)}, RunOutput: int64(3), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": 1.1}, RunOutput: int64(3), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": "1"}, RunOutput: int64(3), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(3)}}},
		{Script: `b = [1, 2]; b[a] = 3`, Input: map[string]interface{}{"a": "a"}, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"b": []interface{}{int64(1), int64(2)}}},

		{Script: `a`, Input: map[string]interface{}{"a": testSliceEmpty}, RunOutput: testSliceEmpty, Output: map[string]interface{}{"a": testSliceEmpty}},
		{Script: `a += []`, Input: map[string]interface{}{"a": testSliceEmpty}, RunOutput: []interface{}(nil), Output: map[string]interface{}{"a": testSliceEmpty}},

		{Script: `a`, Input: map[string]interface{}{"a": testSlice}, RunOutput: testSlice, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: nil, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[0] = 1`, Input: map[string]interface{}{"a": testSlice}, RunOutput: int64(1), Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: int64(1), Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[0] = nil`, Input: map[string]interface{}{"a": testSlice}, RunOutput: nil, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[0]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: nil, Output: map[string]interface{}{"a": testSlice}},

		{Script: `a[1]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: true, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[1] = false`, Input: map[string]interface{}{"a": testSlice}, RunOutput: false, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: false, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[1] = true`, Input: map[string]interface{}{"a": testSlice}, RunOutput: true, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[1]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: true, Output: map[string]interface{}{"a": testSlice}},

		{Script: `a[2]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: int64(1), Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[2] = 2`, Input: map[string]interface{}{"a": testSlice}, RunOutput: int64(2), Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[2]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: int64(2), Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[2] = 1`, Input: map[string]interface{}{"a": testSlice}, RunOutput: int64(1), Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[2]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: int64(1), Output: map[string]interface{}{"a": testSlice}},

		{Script: `a[3]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: 1.1, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[3] = 2.2`, Input: map[string]interface{}{"a": testSlice}, RunOutput: 2.2, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[3]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: 2.2, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[3] = 1.1`, Input: map[string]interface{}{"a": testSlice}, RunOutput: 1.1, Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[3]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: 1.1, Output: map[string]interface{}{"a": testSlice}},

		{Script: `a[4]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: "a", Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[4] = "x"`, Input: map[string]interface{}{"a": testSlice}, RunOutput: "x", Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[4]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: "x", Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[4] = "a"`, Input: map[string]interface{}{"a": testSlice}, RunOutput: "a", Output: map[string]interface{}{"a": testSlice}},
		{Script: `a[4]`, Input: map[string]interface{}{"a": testSlice}, RunOutput: "a", Output: map[string]interface{}{"a": testSlice}},

		{Script: `a[0][0] = true`, Input: map[string]interface{}{"a": []interface{}{[]string{"a"}}}, RunError: fmt.Errorf("type bool cannot be assigned to type string for slice index"), Output: map[string]interface{}{"a": []interface{}{[]string{"a"}}}},
		{Script: `a[0][0] = "a"`, Input: map[string]interface{}{"a": []interface{}{[]bool{true}}}, RunError: fmt.Errorf("type string cannot be assigned to type bool for slice index"), Output: map[string]interface{}{"a": []interface{}{[]bool{true}}}},

		{Script: `a[0][0] = b[0][0]`, Input: map[string]interface{}{"a": []interface{}{[]bool{true}}, "b": []interface{}{[]string{"b"}}}, RunError: fmt.Errorf("type string cannot be assigned to type bool for slice index"), Output: map[string]interface{}{"a": []interface{}{[]bool{true}}}},
		{Script: `b[0][0] = a[0][0]`, Input: map[string]interface{}{"a": []interface{}{[]bool{true}}, "b": []interface{}{[]string{"b"}}}, RunError: fmt.Errorf("type bool cannot be assigned to type string for slice index"), Output: map[string]interface{}{"a": []interface{}{[]bool{true}}}},

		{Script: `a = make([][]bool); a[0] = make([]bool); a[0] = [true, 1]`, RunError: fmt.Errorf("type []interface {} cannot be assigned to type []bool for slice index"), Output: map[string]interface{}{"a": [][]bool{{}}}},
		{Script: `a = make([][]bool); a[0] = make([]bool); a[0] = [true, false]`, RunOutput: []interface{}{true, false}, Output: map[string]interface{}{"a": [][]bool{{true, false}}}},

		{Script: `a = make([][][]bool); a[0] = make([][]bool); a[0][0] = make([]bool); a[0] = [[true, 1]]`, RunError: fmt.Errorf("type []interface {} cannot be assigned to type [][]bool for slice index"), Output: map[string]interface{}{"a": [][][]bool{{{}}}}},
		{Script: `a = make([][][]bool); a[0] = make([][]bool); a[0][0] = make([]bool); a[0] = [[true, false]]`, RunOutput: []interface{}{[]interface{}{true, false}}, Output: map[string]interface{}{"a": [][][]bool{{{true, false}}}}},
	}
	runTests(t, tests, nil, &Options{Debug: true})
}

func TestSlicesAutoAppend(t *testing.T) {
	tests := []Test{
		{Script: `a[2]`, Input: map[string]interface{}{"a": []bool{true, false}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a[2]`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []int32{1, 2}}},
		{Script: `a[2]`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []int64{1, 2}}},
		{Script: `a[2]`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []float32{1.1, 2.2}}},
		{Script: `a[2]`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []float64{1.1, 2.2}}},
		{Script: `a[2]`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []string{"a", "b"}}},

		{Script: `a[2] = true`, Input: map[string]interface{}{"a": []bool{true, false}}, RunOutput: true, Output: map[string]interface{}{"a": []bool{true, false, true}}},
		{Script: `a[2] = 3`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: int64(3), Output: map[string]interface{}{"a": []int32{1, 2, 3}}},
		{Script: `a[2] = 3`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: int64(3), Output: map[string]interface{}{"a": []int64{1, 2, 3}}},
		{Script: `a[2] = 3.3`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: 3.3, Output: map[string]interface{}{"a": []float32{1.1, 2.2, 3.3}}},
		{Script: `a[2] = 3.3`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: 3.3, Output: map[string]interface{}{"a": []float64{1.1, 2.2, 3.3}}},
		{Script: `a[2] = "c"`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: "c", Output: map[string]interface{}{"a": []string{"a", "b", "c"}}},

		{Script: `a[2] = 3.3`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: 3.3, Output: map[string]interface{}{"a": []int32{1, 2, 3}}},
		{Script: `a[2] = 3.3`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: 3.3, Output: map[string]interface{}{"a": []int64{1, 2, 3}}},

		{Script: `a[3] = true`, Input: map[string]interface{}{"a": []bool{true, false}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a[3] = 4`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []int32{1, 2}}},
		{Script: `a[3] = 4`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []int64{1, 2}}},
		{Script: `a[3] = 4.4`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []float32{1.1, 2.2}}},
		{Script: `a[3] = 4.4`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []float64{1.1, 2.2}}},
		{Script: `a[3] = "d"`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []string{"a", "b"}}},

		{Script: `a[2] = nil`, Input: map[string]interface{}{"a": []bool{true, false}}, Output: map[string]interface{}{"a": []bool{true, false, false}}},
		{Script: `a[2] = nil`, Input: map[string]interface{}{"a": []int32{1, 2}}, Output: map[string]interface{}{"a": []int32{1, 2, 0}}},
		{Script: `a[2] = nil`, Input: map[string]interface{}{"a": []int64{1, 2}}, Output: map[string]interface{}{"a": []int64{1, 2, 0}}},
		{Script: `a[2] = nil`, Input: map[string]interface{}{"a": []float32{1.5, 2.5}}, Output: map[string]interface{}{"a": []float32{1.5, 2.5, 0}}},
		{Script: `a[2] = nil`, Input: map[string]interface{}{"a": []float64{1.5, 2.5}}, Output: map[string]interface{}{"a": []float64{1.5, 2.5, 0}}},
		{Script: `a[2] = nil`, Input: map[string]interface{}{"a": []string{"a", "b"}}, Output: map[string]interface{}{"a": []string{"a", "b", ""}}},

		{Script: `a[2] = "a"`, Input: map[string]interface{}{"a": []int16{1, 2}}, RunError: fmt.Errorf("type string cannot be assigned to type int16 for slice index"), Output: map[string]interface{}{"a": []int16{1, 2}}},
		{Script: `a[2] = true`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunError: fmt.Errorf("type bool cannot be assigned to type int64 for slice index"), Output: map[string]interface{}{"a": []int64{1, 2}}},
		{Script: `a[2] = "a"`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunError: fmt.Errorf("type string cannot be assigned to type int64 for slice index"), Output: map[string]interface{}{"a": []int64{1, 2}}},
		{Script: `a[2] = true`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunError: fmt.Errorf("type bool cannot be assigned to type float32 for slice index"), Output: map[string]interface{}{"a": []float32{1.1, 2.2}}},
		{Script: `a[2] = "a"`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunError: fmt.Errorf("type string cannot be assigned to type float64 for slice index"), Output: map[string]interface{}{"a": []float64{1.1, 2.2}}},
		{Script: `a[2] = true`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunError: fmt.Errorf("type bool cannot be assigned to type string for slice index"), Output: map[string]interface{}{"a": []string{"a", "b"}}},
		{Script: `a[2] = 1.1`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunError: fmt.Errorf("type float64 cannot be assigned to type string for slice index"), Output: map[string]interface{}{"a": []string{"a", "b"}}},

		{Script: `a`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: [][]interface{}{}, Output: map[string]interface{}{"a": [][]interface{}{}}},
		{Script: `a[0] = []`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": [][]interface{}{{}}}},
		{Script: `a[0] = [nil]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: []interface{}{nil}, Output: map[string]interface{}{"a": [][]interface{}{{nil}}}},
		{Script: `a[0] = [true]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: []interface{}{true}, Output: map[string]interface{}{"a": [][]interface{}{{true}}}},
		{Script: `a[0] = [1]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": [][]interface{}{{int64(1)}}}},
		{Script: `a[0] = [1.1]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": [][]interface{}{{1.1}}}},
		{Script: `a[0] = ["b"]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: []interface{}{"b"}, Output: map[string]interface{}{"a": [][]interface{}{{"b"}}}},

		{Script: `a[0] = [nil]; a[0][0]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: nil, Output: map[string]interface{}{"a": [][]interface{}{{nil}}}},
		{Script: `a[0] = [true]; a[0][0]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: true, Output: map[string]interface{}{"a": [][]interface{}{{true}}}},
		{Script: `a[0] = [1]; a[0][0]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: int64(1), Output: map[string]interface{}{"a": [][]interface{}{{int64(1)}}}},
		{Script: `a[0] = [1.1]; a[0][0]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: 1.1, Output: map[string]interface{}{"a": [][]interface{}{{1.1}}}},
		{Script: `a[0] = ["b"]; a[0][0]`, Input: map[string]interface{}{"a": [][]interface{}{}}, RunOutput: "b", Output: map[string]interface{}{"a": [][]interface{}{{"b"}}}},

		{Script: `a = make([]bool); a[0] = 1`, RunError: fmt.Errorf("type int64 cannot be assigned to type bool for slice index"), Output: map[string]interface{}{"a": []bool{}}},
		{Script: `a = make([]bool); a[0] = true; a[1] = false`, RunOutput: false, Output: map[string]interface{}{"a": []bool{true, false}}},

		{Script: `a = make([][]bool); a[0] = [true, 1]`, RunError: fmt.Errorf("type []interface {} cannot be assigned to type []bool for slice index"), Output: map[string]interface{}{"a": [][]bool{}}},
		{Script: `a = make([][]bool); a[0] = [true, false]`, RunOutput: []interface{}{true, false}, Output: map[string]interface{}{"a": [][]bool{{true, false}}}},

		{Script: `a = make([][][]bool); a[0] = [[true, 1]]`, RunError: fmt.Errorf("type []interface {} cannot be assigned to type [][]bool for slice index"), Output: map[string]interface{}{"a": [][][]bool{}}},
		{Script: `a = make([][][]bool); a[0] = [[true, false]]`, RunOutput: []interface{}{[]interface{}{true, false}}, Output: map[string]interface{}{"a": [][][]bool{{{true, false}}}}},
	}
	runTests(t, tests, nil, &Options{Debug: true})
}

func TestMakeSlices(t *testing.T) {
	tests := []Test{
		{Script: `make([]foo)`, RunError: fmt.Errorf("undefined type \"foo\"")},

		{Script: `make([]nilT)`, Types: map[string]interface{}{"nilT": nil}, RunError: fmt.Errorf("cannot make type nil")},
		{Script: `make([][]nilT)`, Types: map[string]interface{}{"nilT": nil}, RunError: fmt.Errorf("cannot make type nil")},
		{Script: `make([][][]nilT)`, Types: map[string]interface{}{"nilT": nil}, RunError: fmt.Errorf("cannot make type nil")},

		{Script: `make([]bool, 1++)`, RunError: fmt.Errorf("invalid operation")},
		{Script: `make([]bool, 0, 1++)`, RunError: fmt.Errorf("invalid operation")},

		// spaces and/or newlines
		{Script: `[]`, RunOutput: []interface{}{}},
		{Script: `[ ]`, RunOutput: []interface{}{}},
		{Script: `[
]`, RunOutput: []interface{}{}},
		{Script: `[
 ]`, RunOutput: []interface{}{}},

		{Script: `make(slice2x)`, Types: map[string]interface{}{"slice2x": [][]interface{}{}}, RunOutput: [][]interface{}{}},

		{Script: `make([]bool)`, RunOutput: []bool{}},
		{Script: `make([]int32)`, RunOutput: []int32{}},
		{Script: `make([]int64)`, RunOutput: []int64{}},
		{Script: `make([]float32)`, RunOutput: []float32{}},
		{Script: `make([]float64)`, RunOutput: []float64{}},
		{Script: `make([]string)`, RunOutput: []string{}},

		{Script: `make([]bool, 0)`, RunOutput: []bool{}},
		{Script: `make([]int32, 0)`, RunOutput: []int32{}},
		{Script: `make([]int64, 0)`, RunOutput: []int64{}},
		{Script: `make([]float32, 0)`, RunOutput: []float32{}},
		{Script: `make([]float64, 0)`, RunOutput: []float64{}},
		{Script: `make([]string, 0)`, RunOutput: []string{}},

		{Script: `make([]bool, 2)`, RunOutput: []bool{false, false}},
		{Script: `make([]int32, 2)`, RunOutput: []int32{int32(0), int32(0)}},
		{Script: `make([]int64, 2)`, RunOutput: []int64{int64(0), int64(0)}},
		{Script: `make([]float32, 2)`, RunOutput: []float32{float32(0), float32(0)}},
		{Script: `make([]float64, 2)`, RunOutput: []float64{float64(0), float64(0)}},
		{Script: `make([]string, 2)`, RunOutput: []string{"", ""}},

		{Script: `make([]bool, 0, 2)`, RunOutput: []bool{}},
		{Script: `make([]int32, 0, 2)`, RunOutput: []int32{}},
		{Script: `make([]int64, 0, 2)`, RunOutput: []int64{}},
		{Script: `make([]float32, 0, 2)`, RunOutput: []float32{}},
		{Script: `make([]float64, 0, 2)`, RunOutput: []float64{}},
		{Script: `make([]string, 0, 2)`, RunOutput: []string{}},

		{Script: `make([]bool, 2, 2)`, RunOutput: []bool{false, false}},
		{Script: `make([]int32, 2, 2)`, RunOutput: []int32{int32(0), int32(0)}},
		{Script: `make([]int64, 2, 2)`, RunOutput: []int64{int64(0), int64(0)}},
		{Script: `make([]float32, 2, 2)`, RunOutput: []float32{float32(0), float32(0)}},
		{Script: `make([]float64, 2, 2)`, RunOutput: []float64{float64(0), float64(0)}},
		{Script: `make([]string, 2, 2)`, RunOutput: []string{"", ""}},

		{Script: `a = make([]bool, 0); a += true; a += false`, RunOutput: []bool{true, false}, Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a = make([]int32, 0); a += 1; a += 2`, RunOutput: []int32{int32(1), int32(2)}, Output: map[string]interface{}{"a": []int32{int32(1), int32(2)}}},
		{Script: `a = make([]int64, 0); a += 1; a += 2`, RunOutput: []int64{int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{int64(1), int64(2)}}},
		{Script: `a = make([]float32, 0); a += 1.1; a += 2.2`, RunOutput: []float32{float32(1.1), float32(2.2)}, Output: map[string]interface{}{"a": []float32{float32(1.1), float32(2.2)}}},
		{Script: `a = make([]float64, 0); a += 1.1; a += 2.2`, RunOutput: []float64{1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}}},
		{Script: `a = make([]string, 0); a += "a"; a += "b"`, RunOutput: []string{"a", "b"}, Output: map[string]interface{}{"a": []string{"a", "b"}}},

		{Script: `a = make([]bool, 2); a[0] = true; a[1] = false`, RunOutput: false, Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a = make([]int32, 2); a[0] = 1; a[1] = 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": []int32{int32(1), int32(2)}}},
		{Script: `a = make([]int64, 2); a[0] = 1; a[1] = 2`, RunOutput: int64(2), Output: map[string]interface{}{"a": []int64{int64(1), int64(2)}}},
		{Script: `a = make([]float32, 2); a[0] = 1.1; a[1] = 2.2`, RunOutput: 2.2, Output: map[string]interface{}{"a": []float32{float32(1.1), float32(2.2)}}},
		{Script: `a = make([]float64, 2); a[0] = 1.1; a[1] = 2.2`, RunOutput: 2.2, Output: map[string]interface{}{"a": []float64{1.1, 2.2}}},
		{Script: `a = make([]string, 2); a[0] = "a"; a[1] = "b"`, RunOutput: "b", Output: map[string]interface{}{"a": []string{"a", "b"}}},

		{Script: `make([]boolA)`, Types: map[string]interface{}{"boolA": []bool{}}, RunOutput: [][]bool{}},
		{Script: `make([]int32A)`, Types: map[string]interface{}{"int32A": []int32{}}, RunOutput: [][]int32{}},
		{Script: `make([]int64A)`, Types: map[string]interface{}{"int64A": []int64{}}, RunOutput: [][]int64{}},
		{Script: `make([]float32A)`, Types: map[string]interface{}{"float32A": []float32{}}, RunOutput: [][]float32{}},
		{Script: `make([]float64A)`, Types: map[string]interface{}{"float64A": []float64{}}, RunOutput: [][]float64{}},
		{Script: `make([]stringA)`, Types: map[string]interface{}{"stringA": []string{}}, RunOutput: [][]string{}},

		{Script: `make([]slice)`, Types: map[string]interface{}{"slice": []interface{}{}}, RunOutput: [][]interface{}{}},
		{Script: `a = make([]slice)`, Types: map[string]interface{}{"slice": []interface{}{}}, RunOutput: [][]interface{}{}, Output: map[string]interface{}{"a": [][]interface{}{}}},

		{Script: `make([][]bool)`, RunOutput: [][]bool{}},
		{Script: `make([][]int32)`, RunOutput: [][]int32{}},
		{Script: `make([][]int64)`, RunOutput: [][]int64{}},
		{Script: `make([][]float32)`, RunOutput: [][]float32{}},
		{Script: `make([][]float64)`, RunOutput: [][]float64{}},
		{Script: `make([][]string)`, RunOutput: [][]string{}},

		{Script: `make([][]bool)`, RunOutput: [][]bool{}},
		{Script: `make([][]int32)`, RunOutput: [][]int32{}},
		{Script: `make([][]int64)`, RunOutput: [][]int64{}},
		{Script: `make([][]float32)`, RunOutput: [][]float32{}},
		{Script: `make([][]float64)`, RunOutput: [][]float64{}},
		{Script: `make([][]string)`, RunOutput: [][]string{}},

		{Script: `make([][][]bool)`, RunOutput: [][][]bool{}},
		{Script: `make([][][]int32)`, RunOutput: [][][]int32{}},
		{Script: `make([][][]int64)`, RunOutput: [][][]int64{}},
		{Script: `make([][][]float32)`, RunOutput: [][][]float32{}},
		{Script: `make([][][]float64)`, RunOutput: [][][]float64{}},
		{Script: `make([][][]string)`, RunOutput: [][][]string{}},

		// slice type errors
		{Script: `[]nilT{"a"}`, Types: map[string]interface{}{"nilT": nil}, RunError: fmt.Errorf("cannot make type nil")},
		{Script: `[]int64{1++}`, RunError: fmt.Errorf("invalid operation")},
		{Script: `[]int64{"a"}`, RunError: fmt.Errorf("cannot use type string as type int64 as slice value")},

		// slice type
		{Script: `[]interface{nil}`, RunOutput: []interface{}{nil}},
		{Script: `[]bool{true, false}`, RunOutput: []bool{true, false}},
		{Script: `[]int32{1}`, RunOutput: []int32{int32(1)}},
		{Script: `[]int64{2}`, RunOutput: []int64{int64(2)}},
		{Script: `[]float32{3.5}`, RunOutput: []float32{float32(3.5)}},
		{Script: `[]float64{4.5}`, RunOutput: []float64{4.5}},
		{Script: `[]string{"a"}`, RunOutput: []string{"a"}},
	}
	runTests(t, tests, nil, &Options{Debug: true})
}

func TestSliceOfSlices(t *testing.T) {
	tests := []Test{
		{Script: `a = [1, 2]; a[:]`, ParseError: fmt.Errorf("syntax error")},
		{Script: `(1++)[0:0]`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a = [1, 2]; a[1++:0]`, RunError: fmt.Errorf("invalid operation"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2)}}},
		{Script: `a = [1, 2]; a[0:1++]`, RunError: fmt.Errorf("invalid operation"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2)}}},
		{Script: `a = [1, 2]; a[:0]++`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2)}}},
		{Script: `a = [1, 2]; a[:0]--`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2)}}},

		{Script: `a = [1, 2, 3]; a[nil:2]`, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:nil]`, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[-1:0]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:0]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:1]`, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:2]`, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:3]`, RunOutput: []interface{}{int64(1), int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[1:0]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:1]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:2]`, RunOutput: []interface{}{int64(2)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:3]`, RunOutput: []interface{}{int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[2:1]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[2:2]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[2:3]`, RunOutput: []interface{}{int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[2:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[3:2]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[3:3]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[3:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[4:4]`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[-1:]`, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:]`, RunOutput: []interface{}{int64(1), int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:]`, RunOutput: []interface{}{int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[2:]`, RunOutput: []interface{}{int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[3:]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[4:]`, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[:-1]`, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:0]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:1]`, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:2]`, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:3]`, RunOutput: []interface{}{int64(1), int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:4]`, RunError: fmt.Errorf("index out of range"), RunOutput: nil, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `b[1:2] = 4`, RunError: fmt.Errorf("undefined symbol \"b\"")},
		{Script: `a = [1, 2, 3]; a[1++:2] = 4`, RunError: fmt.Errorf("invalid operation"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:2++] = 4`, RunError: fmt.Errorf("invalid operation"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[nil:2] = 4`, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:nil] = 4`, RunError: fmt.Errorf("index must be a number"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[0:0] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:1] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:4] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:0] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:4] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[-1:] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[0:] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[1:] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[4:] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		{Script: `a = [1, 2, 3]; a[:-1] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:0] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:1] = 4`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},
		{Script: `a = [1, 2, 3]; a[:4] = 4`, RunError: fmt.Errorf("index out of range"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3)}}},

		// slice assigned elem
		{Script: `a[0][0:0] = 4`, Input: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}}}, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}}}},

		{Script: `a = [{"b": "b"}, {"c": "c"}, {"d": "d"}]; a[0:2].a`, RunError: fmt.Errorf("type slice does not support member operation"), Output: map[string]interface{}{"a": []interface{}{map[interface{}]interface{}{"b": "b"}, map[interface{}]interface{}{"c": "c"}, map[interface{}]interface{}{"d": "d"}}}},

		{Script: `a = [{"b": "b"}, {"c": "c"}, {"d": "d"}]; a[0:2]`, RunOutput: []interface{}{map[interface{}]interface{}{"b": "b"}, map[interface{}]interface{}{"c": "c"}}, Output: map[string]interface{}{"a": []interface{}{map[interface{}]interface{}{"b": "b"}, map[interface{}]interface{}{"c": "c"}, map[interface{}]interface{}{"d": "d"}}}},
		{Script: `a = [{"b": "b"}, {"c": "c"}, {"d": "d"}]; a[0:2][0].b`, RunOutput: "b", Output: map[string]interface{}{"a": []interface{}{map[interface{}]interface{}{"b": "b"}, map[interface{}]interface{}{"c": "c"}, map[interface{}]interface{}{"d": "d"}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:3]`, RunOutput: []interface{}{int64(1), int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][1:3]`, RunOutput: []interface{}{int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][2:3]`, RunOutput: []interface{}{int64(3)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][3:3]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:]`, RunOutput: []interface{}{int64(1), int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][1:]`, RunOutput: []interface{}{int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][2:]`, RunOutput: []interface{}{int64(3)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][3:]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:0]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:1]`, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:2]`, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][0:3]`, RunOutput: []interface{}{int64(1), int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][:0]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][:1]`, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][:2]`, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[0][:3]`, RunOutput: []interface{}{int64(1), int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:3]`, RunOutput: []interface{}{int64(4), int64(5), int64(6)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][1:3]`, RunOutput: []interface{}{int64(5), int64(6)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][2:3]`, RunOutput: []interface{}{int64(6)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][3:3]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:]`, RunOutput: []interface{}{int64(4), int64(5), int64(6)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][1:]`, RunOutput: []interface{}{int64(5), int64(6)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][2:]`, RunOutput: []interface{}{int64(6)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][3:]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:0]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:1]`, RunOutput: []interface{}{int64(4)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:2]`, RunOutput: []interface{}{int64(4), int64(5)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][0:3]`, RunOutput: []interface{}{int64(4), int64(5), int64(6)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][:0]`, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][:1]`, RunOutput: []interface{}{int64(4)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][:2]`, RunOutput: []interface{}{int64(4), int64(5)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},
		{Script: `a = [[1, 2, 3], [4, 5, 6]]; a[1][:3]`, RunOutput: []interface{}{int64(4), int64(5), int64(6)}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(4), int64(5), int64(6)}}}},

		{Script: `a = [["123"], ["456"]]; a[0][0][0:3]`, RunOutput: "123", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][1:3]`, RunOutput: "23", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][2:3]`, RunOutput: "3", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][3:3]`, RunOutput: "", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[0][0][0:]`, RunOutput: "123", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][1:]`, RunOutput: "23", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][2:]`, RunOutput: "3", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][3:]`, RunOutput: "", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[0][0][0:0]`, RunOutput: "", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][0:1]`, RunOutput: "1", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][0:2]`, RunOutput: "12", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][0:3]`, RunOutput: "123", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[0][0][:0]`, RunOutput: "", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][:1]`, RunOutput: "1", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][:2]`, RunOutput: "12", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[0][0][:3]`, RunOutput: "123", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[1][0][0:3]`, RunOutput: "456", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][1:3]`, RunOutput: "56", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][2:3]`, RunOutput: "6", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][3:3]`, RunOutput: "", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[1][0][0:]`, RunOutput: "456", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][1:]`, RunOutput: "56", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][2:]`, RunOutput: "6", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][3:]`, RunOutput: "", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[1][0][0:0]`, RunOutput: "", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][0:1]`, RunOutput: "4", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][0:2]`, RunOutput: "45", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][0:3]`, RunOutput: "456", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},

		{Script: `a = [["123"], ["456"]]; a[1][0][:0]`, RunOutput: "", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][:1]`, RunOutput: "4", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][:2]`, RunOutput: "45", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},
		{Script: `a = [["123"], ["456"]]; a[1][0][:3]`, RunOutput: "456", Output: map[string]interface{}{"a": []interface{}{[]interface{}{"123"}, []interface{}{"456"}}}},

		// cap errors
		{Script: `a = "a"; a[0:1:1]`, RunError: fmt.Errorf("type string does not support cap")},
		{Script: `a = [1,2,3,4]; a[1:3:-1]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[1:3:0]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[1:3:2]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[1:3:5]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[:2:-1]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[:2:0]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[:2:1]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[:2:5]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[1:3:"a"]`, RunError: fmt.Errorf("cap must be a number")},
		{Script: `a = [1,2,3,4]; a[1:3:1++]`, RunError: fmt.Errorf("invalid operation")},
		{Script: `a = "a"; b = a[0:1:1]`, RunError: fmt.Errorf("type string does not support cap")},
		{Script: `a = [1,2,3,4]; b = a[1:3:-1]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; b = a[1:3:0]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; b = a[1:3:2]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; b = a[1:3:5]`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; b = a[1:3:"a"]`, RunError: fmt.Errorf("cap must be a number")},
		{Script: `a = [1,2,3,4]; b = a[1:3:1++]`, RunError: fmt.Errorf("invalid operation")},

		// cap assigned errors
		{Script: `a = [1,2,3,4]; a[1:1:-1] = 3`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[1:1:0] = 3`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[1:1:5] = 3`, RunError: fmt.Errorf("cap out of range")},
		{Script: `a = [1,2,3,4]; a[1:1:"a"] = 3`, RunError: fmt.Errorf("cap must be a number")},
		{Script: `a = [1,2,3,4]; a[1:3:1++] = 3`, RunError: fmt.Errorf("invalid operation")},

		// cap
		{Script: `a = [1,2,3,4]; a[1:3:3]`, RunOutput: []interface{}{int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}}},
		{Script: `a = [1,2,3,4]; a[1:3:4]`, RunOutput: []interface{}{int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}}},
		{Script: `a = [1,2,3,4]; a[:2:3]`, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}}},
		{Script: `a = [1,2,3,4]; a[:2:4]`, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}}},
		{Script: `a = [1,2,3,4]; b = a[1:3:3]`, RunOutput: []interface{}{int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}, "b": []interface{}{int64(2), int64(3)}}},
		{Script: `a = [1,2,3,4]; b = a[1:3:4]`, RunOutput: []interface{}{int64(2), int64(3)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}, "b": []interface{}{int64(2), int64(3)}}},
		{Script: `a = [1,2,3,4]; b = a[:2:3]`, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}, "b": []interface{}{int64(1), int64(2)}}},
		{Script: `a = [1,2,3,4]; b = a[:2:4]`, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}, "b": []interface{}{int64(1), int64(2)}}},

		// cap assigned
		{Script: `a = [1,2,3,4]; a[1:1:1] = 3`, RunError: fmt.Errorf("slice cannot be assigned"), Output: map[string]interface{}{"a": []interface{}{int64(1), int64(2), int64(3), int64(4)}}},
	}
	runTests(t, tests, nil, &Options{Debug: true})
}

func TestSliceAppendSlices(t *testing.T) {
	tests := []Test{
		{Script: `a += 1`, Input: map[string]interface{}{"a": []bool{true}}, RunError: fmt.Errorf("invalid type conversion")},
		{Script: `a += 1.1`, Input: map[string]interface{}{"a": []bool{true}}, RunError: fmt.Errorf("invalid type conversion")},
		{Script: `a += "a"`, Input: map[string]interface{}{"a": []bool{true}}, RunError: fmt.Errorf("invalid type conversion")},
		{Script: `a += b`, Input: map[string]interface{}{"a": "a", "b": []bool{true}}, RunError: fmt.Errorf("invalid type conversion")},

		{Script: `a += b`, Input: map[string]interface{}{"a": []bool{true}, "b": []string{"b"}}, RunError: fmt.Errorf("invalid type conversion")},
		{Script: `b += a`, Input: map[string]interface{}{"a": []bool{true}, "b": []string{"b"}}, RunError: fmt.Errorf("invalid type conversion")},

		{Script: `a += nil`, Input: map[string]interface{}{"a": []bool{true}}, RunOutput: []bool{true, false}, Output: map[string]interface{}{"a": []bool{true, false}}},

		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []bool{}}, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []bool{}, "b": []interface{}{}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []bool{true}}, RunOutput: []interface{}{true}, Output: map[string]interface{}{"a": []bool{true}, "b": []interface{}{true}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []bool{true, false}}, RunOutput: []interface{}{true, false}, Output: map[string]interface{}{"a": []bool{true, false}, "b": []interface{}{true, false}}},

		{Script: `b = [true]; b += a`, Input: map[string]interface{}{"a": []bool{}}, RunOutput: []interface{}{true}, Output: map[string]interface{}{"a": []bool{}, "b": []interface{}{true}}},
		{Script: `b = [true]; b += a`, Input: map[string]interface{}{"a": []bool{true}}, RunOutput: []interface{}{true, true}, Output: map[string]interface{}{"a": []bool{true}, "b": []interface{}{true, true}}},
		{Script: `b = [true]; b += a`, Input: map[string]interface{}{"a": []bool{true, false}}, RunOutput: []interface{}{true, true, false}, Output: map[string]interface{}{"a": []bool{true, false}, "b": []interface{}{true, true, false}}},

		{Script: `b = [true, false]; b += a`, Input: map[string]interface{}{"a": []bool{}}, RunOutput: []interface{}{true, false}, Output: map[string]interface{}{"a": []bool{}, "b": []interface{}{true, false}}},
		{Script: `b = [true, false]; b += a`, Input: map[string]interface{}{"a": []bool{true}}, RunOutput: []interface{}{true, false, true}, Output: map[string]interface{}{"a": []bool{true}, "b": []interface{}{true, false, true}}},
		{Script: `b = [true, false]; b += a`, Input: map[string]interface{}{"a": []bool{true, false}}, RunOutput: []interface{}{true, false, true, false}, Output: map[string]interface{}{"a": []bool{true, false}, "b": []interface{}{true, false, true, false}}},

		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []int32{}, "b": []interface{}{}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []int32{1}}, RunOutput: []interface{}{int32(1)}, Output: map[string]interface{}{"a": []int32{1}, "b": []interface{}{int32(1)}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: []interface{}{int32(1), int32(2)}, Output: map[string]interface{}{"a": []int32{1, 2}, "b": []interface{}{int32(1), int32(2)}}},

		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []int32{}, "b": []interface{}{int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []int32{1}}, RunOutput: []interface{}{int64(1), int32(1)}, Output: map[string]interface{}{"a": []int32{1}, "b": []interface{}{int64(1), int32(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: []interface{}{int64(1), int32(1), int32(2)}, Output: map[string]interface{}{"a": []int32{1, 2}, "b": []interface{}{int64(1), int32(1), int32(2)}}},

		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []int32{}, "b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []int32{1}}, RunOutput: []interface{}{int64(1), int64(2), int32(1)}, Output: map[string]interface{}{"a": []int32{1}, "b": []interface{}{int64(1), int64(2), int32(1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: []interface{}{int64(1), int64(2), int32(1), int32(2)}, Output: map[string]interface{}{"a": []int32{1, 2}, "b": []interface{}{int64(1), int64(2), int32(1), int32(2)}}},

		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": []int32{}, "b": []interface{}{1.1}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []int32{1}}, RunOutput: []interface{}{1.1, int32(1)}, Output: map[string]interface{}{"a": []int32{1}, "b": []interface{}{1.1, int32(1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: []interface{}{1.1, int32(1), int32(2)}, Output: map[string]interface{}{"a": []int32{1, 2}, "b": []interface{}{1.1, int32(1), int32(2)}}},

		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []interface{}{1.1, 2.2}, Output: map[string]interface{}{"a": []int32{}, "b": []interface{}{1.1, 2.2}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int32{1}}, RunOutput: []interface{}{1.1, 2.2, int32(1)}, Output: map[string]interface{}{"a": []int32{1}, "b": []interface{}{1.1, 2.2, int32(1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: []interface{}{1.1, 2.2, int32(1), int32(2)}, Output: map[string]interface{}{"a": []int32{1, 2}, "b": []interface{}{1.1, 2.2, int32(1), int32(2)}}},

		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []interface{}{int64(1), 2.2}, Output: map[string]interface{}{"a": []int32{}, "b": []interface{}{int64(1), 2.2}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int32{1}}, RunOutput: []interface{}{int64(1), 2.2, int32(1)}, Output: map[string]interface{}{"a": []int32{1}, "b": []interface{}{int64(1), 2.2, int32(1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: []interface{}{int64(1), 2.2, int32(1), int32(2)}, Output: map[string]interface{}{"a": []int32{1, 2}, "b": []interface{}{int64(1), 2.2, int32(1), int32(2)}}},

		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []int32{}}, RunOutput: []interface{}{1.1, int64(2)}, Output: map[string]interface{}{"a": []int32{}, "b": []interface{}{1.1, int64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []int32{1}}, RunOutput: []interface{}{1.1, int64(2), int32(1)}, Output: map[string]interface{}{"a": []int32{1}, "b": []interface{}{1.1, int64(2), int32(1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []int32{1, 2}}, RunOutput: []interface{}{1.1, int64(2), int32(1), int32(2)}, Output: map[string]interface{}{"a": []int32{1, 2}, "b": []interface{}{1.1, int64(2), int32(1), int32(2)}}},

		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []int64{}, "b": []interface{}{}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []int64{1}}, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []int64{1}, "b": []interface{}{int64(1)}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{1, 2}, "b": []interface{}{int64(1), int64(2)}}},

		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []int64{}, "b": []interface{}{int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []int64{1}}, RunOutput: []interface{}{int64(1), int64(1)}, Output: map[string]interface{}{"a": []int64{1}, "b": []interface{}{int64(1), int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: []interface{}{int64(1), int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{1, 2}, "b": []interface{}{int64(1), int64(1), int64(2)}}},

		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{}, "b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []int64{1}}, RunOutput: []interface{}{int64(1), int64(2), int64(1)}, Output: map[string]interface{}{"a": []int64{1}, "b": []interface{}{int64(1), int64(2), int64(1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: []interface{}{int64(1), int64(2), int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{1, 2}, "b": []interface{}{int64(1), int64(2), int64(1), int64(2)}}},

		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": []int64{}, "b": []interface{}{1.1}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []int64{1}}, RunOutput: []interface{}{1.1, int64(1)}, Output: map[string]interface{}{"a": []int64{1}, "b": []interface{}{1.1, int64(1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: []interface{}{1.1, int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{1, 2}, "b": []interface{}{1.1, int64(1), int64(2)}}},

		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []interface{}{1.1, 2.2}, Output: map[string]interface{}{"a": []int64{}, "b": []interface{}{1.1, 2.2}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int64{1}}, RunOutput: []interface{}{1.1, 2.2, int64(1)}, Output: map[string]interface{}{"a": []int64{1}, "b": []interface{}{1.1, 2.2, int64(1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: []interface{}{1.1, 2.2, int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{1, 2}, "b": []interface{}{1.1, 2.2, int64(1), int64(2)}}},

		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []interface{}{int64(1), 2.2}, Output: map[string]interface{}{"a": []int64{}, "b": []interface{}{int64(1), 2.2}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int64{1}}, RunOutput: []interface{}{int64(1), 2.2, int64(1)}, Output: map[string]interface{}{"a": []int64{1}, "b": []interface{}{int64(1), 2.2, int64(1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: []interface{}{int64(1), 2.2, int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{1, 2}, "b": []interface{}{int64(1), 2.2, int64(1), int64(2)}}},

		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []int64{}}, RunOutput: []interface{}{1.1, int64(2)}, Output: map[string]interface{}{"a": []int64{}, "b": []interface{}{1.1, int64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []int64{1}}, RunOutput: []interface{}{1.1, int64(2), int64(1)}, Output: map[string]interface{}{"a": []int64{1}, "b": []interface{}{1.1, int64(2), int64(1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []int64{1, 2}}, RunOutput: []interface{}{1.1, int64(2), int64(1), int64(2)}, Output: map[string]interface{}{"a": []int64{1, 2}, "b": []interface{}{1.1, int64(2), int64(1), int64(2)}}},

		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []float32{}, "b": []interface{}{}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float32{1}}, RunOutput: []interface{}{float32(1)}, Output: map[string]interface{}{"a": []float32{1}, "b": []interface{}{float32(1)}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float32{1, 2}}, RunOutput: []interface{}{float32(1), float32(2)}, Output: map[string]interface{}{"a": []float32{1, 2}, "b": []interface{}{float32(1), float32(2)}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float32{1.1}}, RunOutput: []interface{}{float32(1.1)}, Output: map[string]interface{}{"a": []float32{1.1}, "b": []interface{}{float32(1.1)}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: []interface{}{float32(1.1), float32(2.2)}, Output: map[string]interface{}{"a": []float32{1.1, 2.2}, "b": []interface{}{float32(1.1), float32(2.2)}}},

		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []float32{}, "b": []interface{}{int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float32{1}}, RunOutput: []interface{}{int64(1), float32(1)}, Output: map[string]interface{}{"a": []float32{1}, "b": []interface{}{int64(1), float32(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float32{1, 2}}, RunOutput: []interface{}{int64(1), float32(1), float32(2)}, Output: map[string]interface{}{"a": []float32{1, 2}, "b": []interface{}{int64(1), float32(1), float32(2)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float32{1.1}}, RunOutput: []interface{}{int64(1), float32(1.1)}, Output: map[string]interface{}{"a": []float32{1.1}, "b": []interface{}{int64(1), float32(1.1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: []interface{}{int64(1), float32(1.1), float32(2.2)}, Output: map[string]interface{}{"a": []float32{1.1, 2.2}, "b": []interface{}{int64(1), float32(1.1), float32(2.2)}}},

		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []float32{}, "b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{1}}, RunOutput: []interface{}{int64(1), int64(2), float32(1)}, Output: map[string]interface{}{"a": []float32{1}, "b": []interface{}{int64(1), int64(2), float32(1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{1, 2}}, RunOutput: []interface{}{int64(1), int64(2), float32(1), float32(2)}, Output: map[string]interface{}{"a": []float32{1, 2}, "b": []interface{}{int64(1), int64(2), float32(1), float32(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{1.1}}, RunOutput: []interface{}{int64(1), int64(2), float32(1.1)}, Output: map[string]interface{}{"a": []float32{1.1}, "b": []interface{}{int64(1), int64(2), float32(1.1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: []interface{}{int64(1), int64(2), float32(1.1), float32(2.2)}, Output: map[string]interface{}{"a": []float32{1.1, 2.2}, "b": []interface{}{int64(1), int64(2), float32(1.1), float32(2.2)}}},

		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": []float32{}, "b": []interface{}{1.1}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float32{1}}, RunOutput: []interface{}{1.1, float32(1)}, Output: map[string]interface{}{"a": []float32{1}, "b": []interface{}{1.1, float32(1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float32{1, 2}}, RunOutput: []interface{}{1.1, float32(1), float32(2)}, Output: map[string]interface{}{"a": []float32{1, 2}, "b": []interface{}{1.1, float32(1), float32(2)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float32{1.1}}, RunOutput: []interface{}{1.1, float32(1.1)}, Output: map[string]interface{}{"a": []float32{1.1}, "b": []interface{}{1.1, float32(1.1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: []interface{}{1.1, float32(1.1), float32(2.2)}, Output: map[string]interface{}{"a": []float32{1.1, 2.2}, "b": []interface{}{1.1, float32(1.1), float32(2.2)}}},

		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []interface{}{1.1, 2.2}, Output: map[string]interface{}{"a": []float32{}, "b": []interface{}{1.1, 2.2}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{1}}, RunOutput: []interface{}{1.1, 2.2, float32(1)}, Output: map[string]interface{}{"a": []float32{1}, "b": []interface{}{1.1, 2.2, float32(1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{1, 2}}, RunOutput: []interface{}{1.1, 2.2, float32(1), float32(2)}, Output: map[string]interface{}{"a": []float32{1, 2}, "b": []interface{}{1.1, 2.2, float32(1), float32(2)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{1.1}}, RunOutput: []interface{}{1.1, 2.2, float32(1.1)}, Output: map[string]interface{}{"a": []float32{1.1}, "b": []interface{}{1.1, 2.2, float32(1.1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: []interface{}{1.1, 2.2, float32(1.1), float32(2.2)}, Output: map[string]interface{}{"a": []float32{1.1, 2.2}, "b": []interface{}{1.1, 2.2, float32(1.1), float32(2.2)}}},

		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []interface{}{int64(1), 2.2}, Output: map[string]interface{}{"a": []float32{}, "b": []interface{}{int64(1), 2.2}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{1}}, RunOutput: []interface{}{int64(1), 2.2, float32(1)}, Output: map[string]interface{}{"a": []float32{1}, "b": []interface{}{int64(1), 2.2, float32(1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{1, 2}}, RunOutput: []interface{}{int64(1), 2.2, float32(1), float32(2)}, Output: map[string]interface{}{"a": []float32{1, 2}, "b": []interface{}{int64(1), 2.2, float32(1), float32(2)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{1.1}}, RunOutput: []interface{}{int64(1), 2.2, float32(1.1)}, Output: map[string]interface{}{"a": []float32{1.1}, "b": []interface{}{int64(1), 2.2, float32(1.1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: []interface{}{int64(1), 2.2, float32(1.1), float32(2.2)}, Output: map[string]interface{}{"a": []float32{1.1, 2.2}, "b": []interface{}{int64(1), 2.2, float32(1.1), float32(2.2)}}},

		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{}}, RunOutput: []interface{}{1.1, int64(2)}, Output: map[string]interface{}{"a": []float32{}, "b": []interface{}{1.1, int64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{1}}, RunOutput: []interface{}{1.1, int64(2), float32(1)}, Output: map[string]interface{}{"a": []float32{1}, "b": []interface{}{1.1, int64(2), float32(1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{1, 2}}, RunOutput: []interface{}{1.1, int64(2), float32(1), float32(2)}, Output: map[string]interface{}{"a": []float32{1, 2}, "b": []interface{}{1.1, int64(2), float32(1), float32(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{1.1}}, RunOutput: []interface{}{1.1, int64(2), float32(1.1)}, Output: map[string]interface{}{"a": []float32{1.1}, "b": []interface{}{1.1, int64(2), float32(1.1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float32{1.1, 2.2}}, RunOutput: []interface{}{1.1, int64(2), float32(1.1), float32(2.2)}, Output: map[string]interface{}{"a": []float32{1.1, 2.2}, "b": []interface{}{1.1, int64(2), float32(1.1), float32(2.2)}}},

		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []float64{}, "b": []interface{}{}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float64{1}}, RunOutput: []interface{}{float64(1)}, Output: map[string]interface{}{"a": []float64{1}, "b": []interface{}{float64(1)}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float64{1, 2}}, RunOutput: []interface{}{float64(1), float64(2)}, Output: map[string]interface{}{"a": []float64{1, 2}, "b": []interface{}{float64(1), float64(2)}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float64{1.1}}, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": []float64{1.1}, "b": []interface{}{1.1}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: []interface{}{1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}, "b": []interface{}{1.1, 2.2}}},

		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []interface{}{int64(1)}, Output: map[string]interface{}{"a": []float64{}, "b": []interface{}{int64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float64{1}}, RunOutput: []interface{}{int64(1), float64(1)}, Output: map[string]interface{}{"a": []float64{1}, "b": []interface{}{int64(1), float64(1)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float64{1, 2}}, RunOutput: []interface{}{int64(1), float64(1), float64(2)}, Output: map[string]interface{}{"a": []float64{1, 2}, "b": []interface{}{int64(1), float64(1), float64(2)}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float64{1.1}}, RunOutput: []interface{}{int64(1), 1.1}, Output: map[string]interface{}{"a": []float64{1.1}, "b": []interface{}{int64(1), 1.1}}},
		{Script: `b = [1]; b += a`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: []interface{}{int64(1), 1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}, "b": []interface{}{int64(1), 1.1, 2.2}}},

		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []interface{}{int64(1), int64(2)}, Output: map[string]interface{}{"a": []float64{}, "b": []interface{}{int64(1), int64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{1}}, RunOutput: []interface{}{int64(1), int64(2), float64(1)}, Output: map[string]interface{}{"a": []float64{1}, "b": []interface{}{int64(1), int64(2), float64(1)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{1, 2}}, RunOutput: []interface{}{int64(1), int64(2), float64(1), float64(2)}, Output: map[string]interface{}{"a": []float64{1, 2}, "b": []interface{}{int64(1), int64(2), float64(1), float64(2)}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{1.1}}, RunOutput: []interface{}{int64(1), int64(2), 1.1}, Output: map[string]interface{}{"a": []float64{1.1}, "b": []interface{}{int64(1), int64(2), 1.1}}},
		{Script: `b = [1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: []interface{}{int64(1), int64(2), 1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}, "b": []interface{}{int64(1), int64(2), 1.1, 2.2}}},

		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []interface{}{1.1}, Output: map[string]interface{}{"a": []float64{}, "b": []interface{}{1.1}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float64{1}}, RunOutput: []interface{}{1.1, float64(1)}, Output: map[string]interface{}{"a": []float64{1}, "b": []interface{}{1.1, float64(1)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float64{1, 2}}, RunOutput: []interface{}{1.1, float64(1), float64(2)}, Output: map[string]interface{}{"a": []float64{1, 2}, "b": []interface{}{1.1, float64(1), float64(2)}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float64{1.1}}, RunOutput: []interface{}{1.1, 1.1}, Output: map[string]interface{}{"a": []float64{1.1}, "b": []interface{}{1.1, 1.1}}},
		{Script: `b = [1.1]; b += a`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: []interface{}{1.1, 1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}, "b": []interface{}{1.1, 1.1, 2.2}}},

		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []interface{}{1.1, 2.2}, Output: map[string]interface{}{"a": []float64{}, "b": []interface{}{1.1, 2.2}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{1}}, RunOutput: []interface{}{1.1, 2.2, float64(1)}, Output: map[string]interface{}{"a": []float64{1}, "b": []interface{}{1.1, 2.2, float64(1)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{1, 2}}, RunOutput: []interface{}{1.1, 2.2, float64(1), float64(2)}, Output: map[string]interface{}{"a": []float64{1, 2}, "b": []interface{}{1.1, 2.2, float64(1), float64(2)}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{1.1}}, RunOutput: []interface{}{1.1, 2.2, 1.1}, Output: map[string]interface{}{"a": []float64{1.1}, "b": []interface{}{1.1, 2.2, 1.1}}},
		{Script: `b = [1.1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: []interface{}{1.1, 2.2, 1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}, "b": []interface{}{1.1, 2.2, 1.1, 2.2}}},

		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []interface{}{int64(1), 2.2}, Output: map[string]interface{}{"a": []float64{}, "b": []interface{}{int64(1), 2.2}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{1}}, RunOutput: []interface{}{int64(1), 2.2, float64(1)}, Output: map[string]interface{}{"a": []float64{1}, "b": []interface{}{int64(1), 2.2, float64(1)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{1, 2}}, RunOutput: []interface{}{int64(1), 2.2, float64(1), float64(2)}, Output: map[string]interface{}{"a": []float64{1, 2}, "b": []interface{}{int64(1), 2.2, float64(1), float64(2)}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{1.1}}, RunOutput: []interface{}{int64(1), 2.2, 1.1}, Output: map[string]interface{}{"a": []float64{1.1}, "b": []interface{}{int64(1), 2.2, 1.1}}},
		{Script: `b = [1, 2.2]; b += a`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: []interface{}{int64(1), 2.2, 1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}, "b": []interface{}{int64(1), 2.2, 1.1, 2.2}}},

		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{}}, RunOutput: []interface{}{1.1, int64(2)}, Output: map[string]interface{}{"a": []float64{}, "b": []interface{}{1.1, int64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{1}}, RunOutput: []interface{}{1.1, int64(2), float64(1)}, Output: map[string]interface{}{"a": []float64{1}, "b": []interface{}{1.1, int64(2), float64(1)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{1, 2}}, RunOutput: []interface{}{1.1, int64(2), float64(1), float64(2)}, Output: map[string]interface{}{"a": []float64{1, 2}, "b": []interface{}{1.1, int64(2), float64(1), float64(2)}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{1.1}}, RunOutput: []interface{}{1.1, int64(2), 1.1}, Output: map[string]interface{}{"a": []float64{1.1}, "b": []interface{}{1.1, int64(2), 1.1}}},
		{Script: `b = [1.1, 2]; b += a`, Input: map[string]interface{}{"a": []float64{1.1, 2.2}}, RunOutput: []interface{}{1.1, int64(2), 1.1, 2.2}, Output: map[string]interface{}{"a": []float64{1.1, 2.2}, "b": []interface{}{1.1, int64(2), 1.1, 2.2}}},

		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []string{}}, RunOutput: []interface{}{}, Output: map[string]interface{}{"a": []string{}, "b": []interface{}{}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []string{"a"}}, RunOutput: []interface{}{"a"}, Output: map[string]interface{}{"a": []string{"a"}, "b": []interface{}{"a"}}},
		{Script: `b = []; b += a`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: []interface{}{"a", "b"}, Output: map[string]interface{}{"a": []string{"a", "b"}, "b": []interface{}{"a", "b"}}},

		{Script: `b = ["a"]; b += a`, Input: map[string]interface{}{"a": []string{}}, RunOutput: []interface{}{"a"}, Output: map[string]interface{}{"a": []string{}, "b": []interface{}{"a"}}},
		{Script: `b = ["a"]; b += a`, Input: map[string]interface{}{"a": []string{"a"}}, RunOutput: []interface{}{"a", "a"}, Output: map[string]interface{}{"a": []string{"a"}, "b": []interface{}{"a", "a"}}},
		{Script: `b = ["a"]; b += a`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: []interface{}{"a", "a", "b"}, Output: map[string]interface{}{"a": []string{"a", "b"}, "b": []interface{}{"a", "a", "b"}}},

		{Script: `b = ["a", "b"]; b += a`, Input: map[string]interface{}{"a": []string{}}, RunOutput: []interface{}{"a", "b"}, Output: map[string]interface{}{"a": []string{}, "b": []interface{}{"a", "b"}}},
		{Script: `b = ["a", "b"]; b += a`, Input: map[string]interface{}{"a": []string{"a"}}, RunOutput: []interface{}{"a", "b", "a"}, Output: map[string]interface{}{"a": []string{"a"}, "b": []interface{}{"a", "b", "a"}}},
		{Script: `b = ["a", "b"]; b += a`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: []interface{}{"a", "b", "a", "b"}, Output: map[string]interface{}{"a": []string{"a", "b"}, "b": []interface{}{"a", "b", "a", "b"}}},

		{Script: `b = [1, "a"]; b += a`, Input: map[string]interface{}{"a": []string{}}, RunOutput: []interface{}{int64(1), "a"}, Output: map[string]interface{}{"a": []string{}, "b": []interface{}{int64(1), "a"}}},
		{Script: `b = [1, "a"]; b += a`, Input: map[string]interface{}{"a": []string{"a"}}, RunOutput: []interface{}{int64(1), "a", "a"}, Output: map[string]interface{}{"a": []string{"a"}, "b": []interface{}{int64(1), "a", "a"}}},
		{Script: `b = [1, "a"]; b += a`, Input: map[string]interface{}{"a": []string{"a", "b"}}, RunOutput: []interface{}{int64(1), "a", "a", "b"}, Output: map[string]interface{}{"a": []string{"a", "b"}, "b": []interface{}{int64(1), "a", "a", "b"}}},

		{Script: `a = []; a += [nil, nil]; a += [nil, nil]`, RunOutput: []interface{}{interface{}(nil), interface{}(nil), interface{}(nil), interface{}(nil)}, Output: map[string]interface{}{"a": []interface{}{interface{}(nil), interface{}(nil), interface{}(nil), interface{}(nil)}}},
		{Script: `a = []; a += [true, false]; a += [false, true]`, RunOutput: []interface{}{interface{}(true), interface{}(false), interface{}(false), interface{}(true)}, Output: map[string]interface{}{"a": []interface{}{interface{}(true), interface{}(false), interface{}(false), interface{}(true)}}},
		{Script: `a = []; a += [1, 2]; a += [3, 4]`, RunOutput: []interface{}{interface{}(int64(1)), interface{}(int64(2)), interface{}(int64(3)), interface{}(int64(4))}, Output: map[string]interface{}{"a": []interface{}{interface{}(int64(1)), interface{}(int64(2)), interface{}(int64(3)), interface{}(int64(4))}}},
		{Script: `a = []; a += [1.1, 2.2]; a += [3.3, 4.4]`, RunOutput: []interface{}{interface{}(1.1), interface{}(2.2), interface{}(3.3), interface{}(4.4)}, Output: map[string]interface{}{"a": []interface{}{interface{}(1.1), interface{}(2.2), interface{}(3.3), interface{}(4.4)}}},
		{Script: `a = []; a += ["a", "b"]; a += ["c", "d"]`, RunOutput: []interface{}{interface{}("a"), interface{}("b"), interface{}("c"), interface{}("d")}, Output: map[string]interface{}{"a": []interface{}{interface{}("a"), interface{}("b"), interface{}("c"), interface{}("d")}}},

		{Script: `a = []; a += [[nil, nil]]; a += [[nil, nil]]`, RunOutput: []interface{}{[]interface{}{nil, nil}, []interface{}{nil, nil}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{nil, nil}, []interface{}{nil, nil}}}},
		{Script: `a = []; a += [[true, false]]; a += [[false, true]]`, RunOutput: []interface{}{[]interface{}{true, false}, []interface{}{false, true}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{true, false}, []interface{}{false, true}}}},
		{Script: `a = []; a += [[1, 2]]; a += [[3, 4]]`, RunOutput: []interface{}{[]interface{}{int64(1), int64(2)}, []interface{}{int64(3), int64(4)}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{int64(1), int64(2)}, []interface{}{int64(3), int64(4)}}}},
		{Script: `a = []; a += [[1.1, 2.2]]; a += [[3.3, 4.4]]`, RunOutput: []interface{}{[]interface{}{1.1, 2.2}, []interface{}{3.3, 4.4}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{1.1, 2.2}, []interface{}{3.3, 4.4}}}},
		{Script: `a = []; a += [["a", "b"]]; a += [["c", "d"]]`, RunOutput: []interface{}{[]interface{}{"a", "b"}, []interface{}{"c", "d"}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{"a", "b"}, []interface{}{"c", "d"}}}},

		{Script: `a = make([]bool); a += 1`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{}}},
		{Script: `a = make([]bool); a += true; a += 1`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{true}}},
		{Script: `a = make([]bool); a += [1]`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{}}},
		{Script: `a = make([]bool); a += [true, 1]`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{}}},

		{Script: `a = make([]bool); a += true; a += false`, RunOutput: []bool{true, false}, Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a = make([]bool); a += [true]; a += [false]`, RunOutput: []bool{true, false}, Output: map[string]interface{}{"a": []bool{true, false}}},
		{Script: `a = make([]bool); a += [true, false]`, RunOutput: []bool{true, false}, Output: map[string]interface{}{"a": []bool{true, false}}},

		{Script: `a = make([]bool); a += true; a += b`, Input: map[string]interface{}{"b": int64(1)}, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{true}, "b": int64(1)}},
		{Script: `a = make([]bool); a += [true]; a += [b]`, Input: map[string]interface{}{"b": int64(1)}, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{true}, "b": int64(1)}},
		{Script: `a = make([]bool); a += [true, b]`, Input: map[string]interface{}{"b": int64(1)}, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{}, "b": int64(1)}},

		{Script: `a = make([]bool); a += b; a += b`, Input: map[string]interface{}{"b": true}, RunOutput: []bool{true, true}, Output: map[string]interface{}{"a": []bool{true, true}, "b": true}},
		{Script: `a = make([]bool); a += [b]; a += [b]`, Input: map[string]interface{}{"b": true}, RunOutput: []bool{true, true}, Output: map[string]interface{}{"a": []bool{true, true}, "b": true}},
		{Script: `a = make([]bool); a += [b, b]`, Input: map[string]interface{}{"b": true}, RunOutput: []bool{true, true}, Output: map[string]interface{}{"a": []bool{true, true}, "b": true}},

		{Script: `a = make([]bool); a += [true, false]; b = make([]int64); b += [1, 2]; a += b`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{true, false}, "b": []int64{int64(1), int64(2)}}},

		{Script: `a = []; b = []; b += true; b += false; a += b`, RunOutput: []interface{}{true, false}, Output: map[string]interface{}{"a": []interface{}{true, false}, "b": []interface{}{true, false}}},
		{Script: `a = []; b = make([]bool); b += true; b += false; a += b`, RunOutput: []interface{}{true, false}, Output: map[string]interface{}{"a": []interface{}{true, false}, "b": []bool{true, false}}},
		{Script: `a = []; b = []; b += [true]; b += [false]; a += [b]`, RunOutput: []interface{}{[]interface{}{true, false}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{true, false}}, "b": []interface{}{true, false}}},
		{Script: `a = []; b = make([]bool); b += [true]; b += [false]; a += [b]`, RunOutput: []interface{}{[]bool{true, false}}, Output: map[string]interface{}{"a": []interface{}{[]bool{true, false}}, "b": []bool{true, false}}},

		{Script: `a = [true, false]; b = [true, false]; a += b`, RunOutput: []interface{}{true, false, true, false}, Output: map[string]interface{}{"a": []interface{}{true, false, true, false}, "b": []interface{}{true, false}}},
		{Script: `a = make([]bool); a += [true, false]; b = make([]bool); b += [true, false]; a += b`, RunOutput: []bool{true, false, true, false}, Output: map[string]interface{}{"a": []bool{true, false, true, false}, "b": []bool{true, false}}},
		{Script: `a = make([]bool); a += [true, false]; b = [true, false]; a += b`, RunOutput: []bool{true, false, true, false}, Output: map[string]interface{}{"a": []bool{true, false, true, false}, "b": []interface{}{true, false}}},
		{Script: `a = [true, false]; b = make([]bool); b += [true, false]; a += b`, RunOutput: []interface{}{true, false, true, false}, Output: map[string]interface{}{"a": []interface{}{true, false, true, false}, "b": []bool{true, false}}},

		{Script: `a = make([][]bool); b = make([][]bool);  a += b`, RunOutput: [][]bool{}, Output: map[string]interface{}{"a": [][]bool{}, "b": [][]bool{}}},
		{Script: `a = make([][]bool); b = make([][]bool); b += [[]]; a += b`, RunOutput: [][]bool{{}}, Output: map[string]interface{}{"a": [][]bool{{}}, "b": [][]bool{{}}}},
		{Script: `a = make([][]bool); a += [[]]; b = make([][]bool); a += b`, RunOutput: [][]bool{{}}, Output: map[string]interface{}{"a": [][]bool{{}}, "b": [][]bool{}}},
		{Script: `a = make([][]bool); a += [[]]; b = make([][]bool); b += [[]]; a += b`, RunOutput: [][]bool{{}, {}}, Output: map[string]interface{}{"a": [][]bool{{}, {}}, "b": [][]bool{{}}}},

		{Script: `a = make([]bool); a += []; b = make([]bool); b += []; a += b`, RunOutput: []bool{}, Output: map[string]interface{}{"a": []bool{}, "b": []bool{}}},
		{Script: `a = make([]bool); a += [true]; b = make([]bool); b += []; a += b`, RunOutput: []bool{true}, Output: map[string]interface{}{"a": []bool{true}, "b": []bool{}}},
		{Script: `a = make([]bool); a += []; b = make([]bool); b += [true]; a += b`, RunOutput: []bool{true}, Output: map[string]interface{}{"a": []bool{true}, "b": []bool{true}}},
		{Script: `a = make([]bool); a += [true]; b = make([]bool); b += [true]; a += b`, RunOutput: []bool{true, true}, Output: map[string]interface{}{"a": []bool{true, true}, "b": []bool{true}}},

		{Script: `a = make([][]bool); a += [true, false];`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": [][]bool{}}},
		{Script: `a = make([][]bool); a += [[true, false]]; b = make([]bool); b += [true, false]; a += b`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": [][]bool{{true, false}}, "b": []bool{true, false}}},
		{Script: `a = make([][]bool); a += [[true, false]]; b = [[1.1]]; a += b`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": [][]bool{{true, false}}, "b": []interface{}{[]interface{}{1.1}}}},
		{Script: `a = make([]bool); a += [true, false]; b = make([][]bool); b += [[true, false]]; a += b`, RunError: fmt.Errorf("invalid type conversion"), Output: map[string]interface{}{"a": []bool{true, false}, "b": [][]bool{{true, false}}}},

		{Script: `a = make([][]interface); a += [[1, 2]]`, RunOutput: [][]interface{}{{int64(1), int64(2)}}, Output: map[string]interface{}{"a": [][]interface{}{{int64(1), int64(2)}}}},
		{Script: `a = make([][]interface); b = [1, 2]; a += [b]`, RunOutput: [][]interface{}{{int64(1), int64(2)}}, Output: map[string]interface{}{"a": [][]interface{}{{int64(1), int64(2)}}, "b": []interface{}{int64(1), int64(2)}}},
		{Script: `a = make([][]interface); a += [[1, 2],[3, 4]]`, RunOutput: [][]interface{}{{int64(1), int64(2)}, {int64(3), int64(4)}}, Output: map[string]interface{}{"a": [][]interface{}{{int64(1), int64(2)}, {int64(3), int64(4)}}}},
		{Script: `a = make([][]interface); b = [1, 2]; a += [b]; b = [3, 4]; a += [b]`, RunOutput: [][]interface{}{{int64(1), int64(2)}, {int64(3), int64(4)}}, Output: map[string]interface{}{"a": [][]interface{}{{int64(1), int64(2)}, {int64(3), int64(4)}}, "b": []interface{}{int64(3), int64(4)}}},

		{Script: `a = [["a", "b"], ["c", "d"]]; b = [["w", "x"], ["y", "z"]]; a += b`, RunOutput: []interface{}{[]interface{}{"a", "b"}, []interface{}{"c", "d"}, []interface{}{"w", "x"}, []interface{}{"y", "z"}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{"a", "b"}, []interface{}{"c", "d"}, []interface{}{"w", "x"}, []interface{}{"y", "z"}}, "b": []interface{}{[]interface{}{"w", "x"}, []interface{}{"y", "z"}}}},
		{Script: `a = make([][]string); a += [["a", "b"], ["c", "d"]]; b = make([][]string); b += [["w", "x"], ["y", "z"]]; a += b`, RunOutput: [][]string{{"a", "b"}, {"c", "d"}, {"w", "x"}, {"y", "z"}}, Output: map[string]interface{}{"a": [][]string{{"a", "b"}, {"c", "d"}, {"w", "x"}, {"y", "z"}}, "b": [][]string{{"w", "x"}, {"y", "z"}}}},
		{Script: `a = make([][]string); a += [["a", "b"], ["c", "d"]]; b = [["w", "x"], ["y", "z"]]; a += b`, RunOutput: [][]string{{"a", "b"}, {"c", "d"}, {"w", "x"}, {"y", "z"}}, Output: map[string]interface{}{"a": [][]string{{"a", "b"}, {"c", "d"}, {"w", "x"}, {"y", "z"}}, "b": []interface{}{[]interface{}{"w", "x"}, []interface{}{"y", "z"}}}},
		{Script: `a = [["a", "b"], ["c", "d"]]; b = make([][]string); b += [["w", "x"], ["y", "z"]]; a += b`, RunOutput: []interface{}{[]interface{}{"a", "b"}, []interface{}{"c", "d"}, []string{"w", "x"}, []string{"y", "z"}}, Output: map[string]interface{}{"a": []interface{}{[]interface{}{"a", "b"}, []interface{}{"c", "d"}, []string{"w", "x"}, []string{"y", "z"}}, "b": [][]string{{"w", "x"}, {"y", "z"}}}},

		{Script: `a = make([][]int64); a += [[1, 2], [3, 4]]; b = make([][]int32); b += [[5, 6], [7, 8]]; a += b`, RunOutput: [][]int64{{int64(1), int64(2)}, {int64(3), int64(4)}, {int64(5), int64(6)}, {int64(7), int64(8)}}, Output: map[string]interface{}{"a": [][]int64{{int64(1), int64(2)}, {int64(3), int64(4)}, {int64(5), int64(6)}, {int64(7), int64(8)}}, "b": [][]int32{{int32(5), int32(6)}, {int32(7), int32(8)}}}},
		{Script: `a = make([][]int32); a += [[1, 2], [3, 4]]; b = make([][]int64); b += [[5, 6], [7, 8]]; a += b`, RunOutput: [][]int32{{int32(1), int32(2)}, {int32(3), int32(4)}, {int32(5), int32(6)}, {int32(7), int32(8)}}, Output: map[string]interface{}{"a": [][]int32{{int32(1), int32(2)}, {int32(3), int32(4)}, {int32(5), int32(6)}, {int32(7), int32(8)}}, "b": [][]int64{{int64(5), int64(6)}, {int64(7), int64(8)}}}},
		{Script: `a = make([][]int64); a += [[1, 2], [3, 4]]; b = make([][]float64); b += [[5, 6], [7, 8]]; a += b`, RunOutput: [][]int64{{int64(1), int64(2)}, {int64(3), int64(4)}, {int64(5), int64(6)}, {int64(7), int64(8)}}, Output: map[string]interface{}{"a": [][]int64{{int64(1), int64(2)}, {int64(3), int64(4)}, {int64(5), int64(6)}, {int64(7), int64(8)}}, "b": [][]float64{{float64(5), float64(6)}, {float64(7), float64(8)}}}},
		{Script: `a = make([][]float64); a += [[1, 2], [3, 4]]; b = make([][]int64); b += [[5, 6], [7, 8]]; a += b`, RunOutput: [][]float64{{float64(1), float64(2)}, {float64(3), float64(4)}, {float64(5), float64(6)}, {float64(7), float64(8)}}, Output: map[string]interface{}{"a": [][]float64{{float64(1), float64(2)}, {float64(3), float64(4)}, {float64(5), float64(6)}, {float64(7), float64(8)}}, "b": [][]int64{{int64(5), int64(6)}, {int64(7), int64(8)}}}},

		{Script: `a = make([][][]interface); a += [[[1, 2]]]`, RunOutput: [][][]interface{}{{{int64(1), int64(2)}}}, Output: map[string]interface{}{"a": [][][]interface{}{{{int64(1), int64(2)}}}}},
		{Script: `a = make([][][]interface); b = [[1, 2]]; a += [b]`, RunOutput: [][][]interface{}{{{int64(1), int64(2)}}}, Output: map[string]interface{}{"a": [][][]interface{}{{{int64(1), int64(2)}}}, "b": []interface{}{[]interface{}{int64(1), int64(2)}}}},
		{Script: `a = make([][][]interface); b = [1, 2]; a += [[b]]`, RunOutput: [][][]interface{}{{{int64(1), int64(2)}}}, Output: map[string]interface{}{"a": [][][]interface{}{{{int64(1), int64(2)}}}, "b": []interface{}{int64(1), int64(2)}}},

		{Script: `a = make([][][]interface); a += [[[1, 2],[3, 4]]]`, RunOutput: [][][]interface{}{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, Output: map[string]interface{}{"a": [][][]interface{}{{{int64(1), int64(2)}, {int64(3), int64(4)}}}}},
		{Script: `a = make([][][]interface); b = [[1, 2],[3, 4]]; a += [b]`, RunOutput: [][][]interface{}{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, Output: map[string]interface{}{"a": [][][]interface{}{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, "b": []interface{}{[]interface{}{int64(1), int64(2)}, []interface{}{int64(3), int64(4)}}}},
		{Script: `a = make([][][]interface); b = [1, 2]; c = [b]; b = [3, 4]; c += [b]; a += [c]`, RunOutput: [][][]interface{}{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, Output: map[string]interface{}{"a": [][][]interface{}{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, "b": []interface{}{int64(3), int64(4)}, "c": []interface{}{[]interface{}{int64(1), int64(2)}, []interface{}{int64(3), int64(4)}}}},
		{Script: `a = make([][][]interface); b = [1, 2]; c = []; c += [b]; b = [3, 4]; c += [b]; a += [c]`, RunOutput: [][][]interface{}{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, Output: map[string]interface{}{"a": [][][]interface{}{{{int64(1), int64(2)}, {int64(3), int64(4)}}}, "b": []interface{}{int64(3), int64(4)}, "c": []interface{}{[]interface{}{int64(1), int64(2)}, []interface{}{int64(3), int64(4)}}}},
	}
	runTests(t, tests, nil, &Options{Debug: true})
}

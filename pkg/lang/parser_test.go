package lang

import (
	"strings"
	"testing"
)

func TestArithmeticIndexExpressionParsesInsideCallChain(t *testing.T) {
	_, err := Parse(strings.NewReader("fn test() { split = url.split(\"/\"); return split[split.length() - 3] }"))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
}

func TestArithmeticIndexExpressionParsesAndEvaluates(t *testing.T) {
	env, err := Execute(strings.NewReader("xs = [\"a\", \"b\", \"c\", \"d\"]\nout = xs[3 - 1]\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	out, getErr := env.Get("out")
	if getErr != nil {
		t.Fatalf("missing out value: %v", getErr)
	}

	str, ok := out.ToString()
	if !ok {
		t.Fatalf("out is not a string: %#v", out)
	}
	if str.String() != "c" {
		t.Fatalf("unexpected out value: got %q want %q", str.String(), "c")
	}
}

func TestArithmeticPrecedenceEvaluatesCorrectly(t *testing.T) {
	env, err := Execute(strings.NewReader("value = 1 + 2 * 3 - 4 / 2\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	value, getErr := env.Get("value")
	if getErr != nil {
		t.Fatalf("missing value: %v", getErr)
	}

	num, ok := value.ToNumber()
	if !ok {
		t.Fatalf("value is not numeric: %#v", value)
	}
	if num != 5 {
		t.Fatalf("unexpected value: got %v want %v", num, 5)
	}
}

func TestUnaryMinusEvaluatesCorrectly(t *testing.T) {
	env, err := Execute(strings.NewReader("value = -3 + 5\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	value, getErr := env.Get("value")
	if getErr != nil {
		t.Fatalf("missing value: %v", getErr)
	}

	num, ok := value.ToNumber()
	if !ok {
		t.Fatalf("value is not numeric: %#v", value)
	}
	if num != 2 {
		t.Fatalf("unexpected value: got %v want %v", num, 2)
	}
}

func TestLogicalNotEvaluatesCorrectly(t *testing.T) {
	env, err := Execute(strings.NewReader("value = !false\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	value, getErr := env.Get("value")
	if getErr != nil {
		t.Fatalf("missing value: %v", getErr)
	}

	b, ok := value.ToBool()
	if !ok {
		t.Fatalf("value is not boolean: %#v", value)
	}
	if !b {
		t.Fatalf("unexpected value: got %v want %v", b, true)
	}
}

func TestNotEqualsEvaluatesCorrectly(t *testing.T) {
	env, err := Execute(strings.NewReader("value = 1 != 2\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	value, getErr := env.Get("value")
	if getErr != nil {
		t.Fatalf("missing value: %v", getErr)
	}

	b, ok := value.ToBool()
	if !ok {
		t.Fatalf("value is not boolean: %#v", value)
	}
	if !b {
		t.Fatalf("unexpected value: got %v want %v", b, true)
	}
}

func TestBreakExitsLoopEarly(t *testing.T) {
	env, err := Execute(strings.NewReader("count = 0\nfor item in [1, 2, 3, 4] {\n    count = count + 1\n    if item == 2 {\n        break\n    }\n}\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	count, getErr := env.Get("count")
	if getErr != nil {
		t.Fatalf("missing count: %v", getErr)
	}

	num, ok := count.ToNumber()
	if !ok {
		t.Fatalf("count is not numeric: %#v", count)
	}
	if num != 2 {
		t.Fatalf("unexpected count: got %v want %v", num, 2)
	}
}

func TestContinueSkipsCurrentIteration(t *testing.T) {
	env, err := Execute(strings.NewReader("count = 0\nfor item in [1, 2, 3, 4] {\n    if item == 2 {\n        continue\n    }\n    count = count + item\n}\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	count, getErr := env.Get("count")
	if getErr != nil {
		t.Fatalf("missing count: %v", getErr)
	}

	num, ok := count.ToNumber()
	if !ok {
		t.Fatalf("count is not numeric: %#v", count)
	}
	if num != 8 {
		t.Fatalf("unexpected count: got %v want %v", num, 8)
	}
}

func TestNestedBreakOnlyExitsInnerLoop(t *testing.T) {
	env, err := Execute(strings.NewReader("count = 0\nfor outer in [1, 2, 3] {\n    for inner in [10, 20, 30] {\n        count = count + 1\n        if inner == 20 {\n            break\n        }\n    }\n}\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	count, getErr := env.Get("count")
	if getErr != nil {
		t.Fatalf("missing count: %v", getErr)
	}

	num, ok := count.ToNumber()
	if !ok {
		t.Fatalf("count is not numeric: %#v", count)
	}
	if num != 6 {
		t.Fatalf("unexpected count: got %v want %v", num, 6)
	}
}

func TestNestedContinueOnlySkipsInnerIteration(t *testing.T) {
	env, err := Execute(strings.NewReader("count = 0\nfor outer in [1, 2] {\n    for inner in [10, 20, 30] {\n        if inner == 20 {\n            continue\n        }\n        count = count + 1\n    }\n}\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	count, getErr := env.Get("count")
	if getErr != nil {
		t.Fatalf("missing count: %v", getErr)
	}

	num, ok := count.ToNumber()
	if !ok {
		t.Fatalf("count is not numeric: %#v", count)
	}
	if num != 4 {
		t.Fatalf("unexpected count: got %v want %v", num, 4)
	}
}

func TestBreakOutsideLoopErrors(t *testing.T) {
	_, err := Execute(strings.NewReader("break\n"))
	if err == nil {
		t.Fatal("expected break outside loop to fail")
	}
	if err.Error() != "break not allowed in this context" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContinueOutsideLoopErrors(t *testing.T) {
	_, err := Execute(strings.NewReader("continue\n"))
	if err == nil {
		t.Fatal("expected continue outside loop to fail")
	}
	if err.Error() != "continue not allowed in this context" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReturnInsideLoopEscapesFunction(t *testing.T) {
	env, err := Execute(strings.NewReader("fn test() {\n    for item in [1, 2, 3] {\n        if item == 2 {\n            return item\n        }\n    }\n    return 0\n}\n"))
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	testFn, getErr := env.Get("test")
	if getErr != nil {
		t.Fatalf("missing test: %v", getErr)
	}

	out, callErr := testFn.Call()
	if callErr != nil {
		t.Fatalf("call failed: %v", callErr)
	}

	num, ok := out.ToNumber()
	if !ok {
		t.Fatalf("out is not numeric: %#v", out)
	}
	if num != 2 {
		t.Fatalf("unexpected out: got %v want %v", num, 2)
	}
}

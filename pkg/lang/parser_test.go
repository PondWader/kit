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

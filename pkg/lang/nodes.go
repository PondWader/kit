package lang

import (
	"cmp"
	"fmt"
	"math"
	"strings"

	"github.com/PondWader/kit/pkg/lang/std"
	"github.com/PondWader/kit/pkg/lang/values"
)

type Node interface {
	Eval(*Environment) (values.Value, *values.Error)
	String() string
}

type NodeExport struct {
	Decl NodeDeclaration
}

func (n NodeExport) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Decl.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	e.Export(n.Decl.Name, v)
	return v, nil
}

func (n NodeExport) String() string {
	return fmt.Sprintf("export %s", n.Decl.String())
}

type NodeImport struct {
	Modules []string
}

func (n NodeImport) Eval(e *Environment) (values.Value, *values.Error) {
	for _, modName := range n.Modules {
		if err := e.Import(modName); err != nil {
			return values.Nil, err
		}
	}
	return values.Nil, nil
}

func (n NodeImport) String() string {
	quoted := make([]string, len(n.Modules))
	for i, mod := range n.Modules {
		quoted[i] = mod
	}
	return "import " + strings.Join(quoted, ",")
}

type NodeInterfaceDecl struct {
	Name    string
	Fields  map[string]values.Kind
	Methods []string
}

func (n NodeInterfaceDecl) Eval(e *Environment) (values.Value, *values.Error) {
	iface := values.NewInterface(n.Name)
	for key, kind := range n.Fields {
		iface.AddField(key, kind)
	}
	for _, method := range n.Methods {
		iface.AddMethod(method)
	}
	v := iface.Val()
	e.SetScoped(n.Name, v)
	return v, nil
}

func (n NodeInterfaceDecl) String() string {
	var body []string
	for key, kind := range n.Fields {
		body = append(body, fmt.Sprintf("%s: %s", key, kind.String()))
	}
	for _, method := range n.Methods {
		body = append(body, fmt.Sprintf("fn %s()", method))
	}
	return fmt.Sprintf("interface %s { %s }", n.Name, strings.Join(body, "; "))
}

type NodeDeclaration struct {
	Name  string
	Value Node
}

func (n NodeDeclaration) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Value.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	e.Set(n.Name, v)
	return v, nil
}

func (n NodeDeclaration) String() string {
	return fmt.Sprintf("%s = %s", n.Name, n.Value.String())
}

type NodeList struct {
	Elements []Node
}

func (n NodeList) Eval(e *Environment) (values.Value, *values.Error) {
	list := values.NewList(len(n.Elements))
	for i, el := range n.Elements {
		el, err := el.Eval(e)
		if err != nil {
			return values.Nil, err
		}
		list.Set(i, el)
	}
	return list.Val(), nil
}

func (n NodeList) String() string {
	var parts []string
	for _, el := range n.Elements {
		parts = append(parts, el.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

type NodeObject struct {
	Body []Node
}

func (n NodeObject) Eval(e *Environment) (values.Value, *values.Error) {
	child := e.NewChild()
	child.VarBoundary = true
	if err := child.Execute(n.Body); err != nil {
		return values.Nil, err
	}
	obj := values.ObjectFromMap(child.Vars)
	return obj.Val(), nil
}

func (n NodeObject) String() string {
	var parts []string
	for _, node := range n.Body {
		parts = append(parts, node.String())
	}
	if len(parts) > 0 {
		return fmt.Sprintf("{ %s }", strings.Join(parts, "; "))
	}
	return "{}"
}

type NodeInterfaceInstantiate struct {
	Interface Node
	Value     NodeObject
}

func (n NodeInterfaceInstantiate) Eval(e *Environment) (values.Value, *values.Error) {
	ifaceV, err := n.Interface.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	iface, ok := ifaceV.ToInterface()
	if !ok {
		return values.Nil, values.NewError("expected interface value before object literal")
	}

	v, err := n.Value.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	obj, ok := v.ToObject()
	if !ok {
		return values.Nil, values.NewError("expected object value when creating interface instance")
	}

	if err = iface.ValidateObject(obj); err != nil {
		return values.Nil, err
	}
	obj.TagInterface(iface)

	return v, nil
}

func (n NodeInterfaceInstantiate) String() string {
	return fmt.Sprintf("%s %s", n.Interface.String(), n.Value.String())
}

type NodeBlock struct {
	Body           []Node
	IsFunctionBody bool
}

func (n NodeBlock) Eval(e *Environment) (values.Value, *values.Error) {
	child := e.NewChild()
	if n.IsFunctionBody {
		child.control = NewExec()
		child.control.ReturnAllowed = true
	}
	if err := child.Execute(n.Body); err != nil {
		return values.Nil, err
	}
	if n.IsFunctionBody {
		return child.control.ReturnVal, nil
	}
	return values.Nil, nil
}

func (n NodeBlock) String() string {
	var parts []string
	for _, node := range n.Body {
		parts = append(parts, node.String())
	}
	if len(parts) > 0 {
		return fmt.Sprintf("{ %s }", strings.Join(parts, "; "))
	}
	return "{}"
}

type NodeLiteral struct {
	Value values.Value
}

func (n NodeLiteral) Eval(e *Environment) (values.Value, *values.Error) {
	return n.Value, nil
}

func (n NodeLiteral) String() string {
	return fmt.Sprintf("%v", n.Value)
}

type NodeIdentifier struct {
	Ident string
}

func (n NodeIdentifier) Eval(e *Environment) (values.Value, *values.Error) {
	return e.Get(n.Ident)
}

func (n NodeIdentifier) String() string {
	return n.Ident
}

type NodeString struct {
	Parts []Node
}

func (n NodeString) Eval(e *Environment) (values.Value, *values.Error) {
	// Fast-path for single part
	if len(n.Parts) == 1 {
		v, err := n.Parts[0].Eval(e)
		if err != nil {
			return values.Nil, err
		}
		return v.Stringify().Val(), nil
	}

	var sb strings.Builder
	for _, part := range n.Parts {
		v, err := part.Eval(e)
		if err != nil {
			return values.Nil, err
		}
		sb.WriteString(v.String())
	}
	return values.Of(sb.String()), nil
}

func (n NodeString) String() string {
	if len(n.Parts) == 1 {
		return fmt.Sprintf("\"%s\"", n.Parts[0].String())
	}
	return fmt.Sprintf("\"<string with %d parts>\"", len(n.Parts))
}

type NodeCall struct {
	Fn  Node
	Arg Node
}

func (n NodeCall) Eval(e *Environment) (values.Value, *values.Error) {
	fn, err := n.Fn.Eval(e)
	if err != nil {
		return values.Nil, err
	} else if n.Arg == nil {
		return fn.Call()
	}

	arg, err := n.Arg.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	return fn.Call(arg)
}

func (n NodeCall) String() string {
	if n.Arg == nil {
		return fmt.Sprintf("%s()", n.Fn.String())
	}
	return fmt.Sprintf("%s(%s)", n.Fn.String(), n.Arg.String())
}

type NodeKeyAccess struct {
	Val Node
	Key string
}

func (n NodeKeyAccess) Eval(e *Environment) (values.Value, *values.Error) {
	val, err := n.Val.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	return val.Get(n.Key)
}

func (n NodeKeyAccess) String() string {
	return fmt.Sprintf("%s.%s", n.Val.String(), n.Key)
}

type NodeIndexAccess struct {
	Val   Node
	Index Node
}

func (n NodeIndexAccess) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Val.Eval(e)
	if err != nil {
		return values.Nil, err
	}

	idx, err := n.Index.Eval(e)
	if err != nil {
		return values.Nil, err
	}

	idxNum, ok := idx.ToNumber()
	if !ok {
		return values.Nil, values.NewError("index value must be a number")
	}
	if math.Trunc(idxNum) != idxNum {
		return values.Nil, values.NewError("index value must be a valid integer")
	}

	return v.Index(int(idxNum))
}

func (n NodeIndexAccess) String() string {
	return fmt.Sprintf("%s[%s]", n.Val.String(), n.Index.String())
}

type NodeFunction struct {
	ArgName        string
	ArgKind        values.Kind
	ArgDestructure []string
	Body           Node
}

func (n NodeFunction) Eval(e *Environment) (values.Value, *values.Error) {
	if len(n.ArgDestructure) > 0 {
		return values.Of(func(arg values.Value) (values.Value, *values.Error) {
			obj, ok := arg.ToObject()
			if !ok {
				return values.Nil, values.NewError("expected object as function argument")
			}

			c := e.NewChild()
			for _, argName := range n.ArgDestructure {
				field := obj.Get(argName)
				if field == values.Nil {
					return values.Nil, values.NewError("missing key \"" + argName + "\" in function argument object")
				}
				c.SetScoped(argName, field)
			}
			return n.Body.Eval(c)
		}), nil
	}

	if n.ArgName != "" {
		return values.Of(func(arg values.Value) (values.Value, *values.Error) {
			if n.ArgKind != values.KindUnknownKind && arg.Kind() != n.ArgKind {
				return values.Nil, values.NewError("expected " + n.ArgKind.String() + " as function argument")
			}

			c := e.NewChild()
			c.SetScoped(n.ArgName, arg)
			return n.Body.Eval(c)
		}), nil
	}
	return values.Of(func() (values.Value, *values.Error) {
		return n.Body.Eval(e)
	}), nil
}

func (n NodeFunction) String() string {
	if len(n.ArgDestructure) > 0 {
		return fmt.Sprintf("fn({%s}) %s", strings.Join(n.ArgDestructure, ", "), n.Body.String())
	}

	if n.ArgName != "" {
		if n.ArgKind != values.KindUnknownKind {
			return fmt.Sprintf("fn(%s: %s) %s", n.ArgName, n.ArgKind.String(), n.Body.String())
		}
		return fmt.Sprintf("fn(%s) %s", n.ArgName, n.Body.String())
	}
	return fmt.Sprintf("fn() %s", n.Body.String())
}

type NodeReturn struct {
	Val Node
}

func (n NodeReturn) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Val.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	return v, e.Return(v)
}

func (n NodeReturn) String() string {
	return "return " + n.Val.String()
}

type NodeThrow struct {
	Val Node
}

func (n NodeThrow) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Val.Eval(e)
	if err != nil {
		return values.Nil, err
	}

	obj, ok := v.ToObject()
	if !ok {
		return values.Nil, values.NewError("throw argument must be an instance of Error")
	}

	errorIface, _ := std.Error.ToInterface()
	if !obj.Implements(errorIface) {
		return values.Nil, values.NewError("throw argument must be an instance of Error")
	}

	message, ok := obj.Get("message").ToString()
	if !ok {
		return values.Nil, values.NewError("throw argument must be an instance of Error")
	}

	return values.Nil, values.NewError(message.String())
}

func (n NodeThrow) String() string {
	return "throw " + n.Val.String()
}

type NodeIf struct {
	Condition Node
	Body      Node
	Else      Node
}

func (n NodeIf) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Condition.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	b, ok := v.ToBool()
	if !ok {
		return values.Nil, values.NewError("expected boolean type for if condition")
	}
	if b {
		return n.Body.Eval(e)
	} else if n.Else != nil {
		return n.Else.Eval(e)
	}
	return values.Nil, nil
}

func (n NodeIf) String() string {
	s := "if " + n.Condition.String() + " " + n.Body.String()
	if n.Else != nil {
		s += " else " + n.Else.String()
	}
	return s
}

type NodeEquals struct {
	Left  Node
	Right Node
}

func (n NodeEquals) Eval(e *Environment) (values.Value, *values.Error) {
	left, err := n.Left.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	right, err := n.Right.Eval(e)
	if err != nil {
		return values.Nil, err
	}

	result, err := left.Equals(right)
	return values.Of(result), err
}

func (n NodeEquals) String() string {
	return n.Left.String() + " == " + n.Right.String()
}

type NodeNumberComparison struct {
	Target    int
	Inclusive bool
	Left      Node
	Right     Node
}

func (n NodeNumberComparison) Eval(e *Environment) (values.Value, *values.Error) {
	leftV, err := n.Left.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	rightV, err := n.Right.Eval(e)
	if err != nil {
		return values.Nil, err
	}

	left, ok := leftV.ToNumber()
	if !ok {
		return values.Nil, values.FmtTypeError(n.OpSymbol(), values.KindNumber)
	}
	right, ok := rightV.ToNumber()
	if !ok {
		return values.Nil, values.FmtTypeError(n.OpSymbol(), values.KindNumber)
	}

	return values.Of(cmp.Compare(left, right) == n.Target || (n.Inclusive && left == right)), nil
}

func (n NodeNumberComparison) String() string {
	return n.Left.String() + " " + n.OpSymbol() + " " + n.Right.String()
}

func (n NodeNumberComparison) OpSymbol() string {
	var op string
	if n.Target == 0 {
		op = "=="
	} else if n.Target < 0 {
		op = "<"
	} else if n.Target > 0 {
		op = ">"
	}
	if n.Inclusive && n.Target != 0 {
		op += "="
	}
	return op
}

type NodeNot struct {
	Inner Node
}

func (n NodeNot) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Inner.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	b, ok := v.ToBool()
	if !ok {
		return values.Nil, values.FmtTypeError("! (not)", values.KindBool)
	}
	return values.Of(!b), nil
}

func (n NodeNot) String() string {
	return "!" + n.Inner.String()
}

type LogicalOp uint8

const (
	LogicalOpAnd LogicalOp = iota
	LogicalOpOr
)

func (op LogicalOp) String() string {
	switch op {
	case LogicalOpAnd:
		return "&&"
	case LogicalOpOr:
		return "||"
	default:
		return "<invalid op>"
	}
}

type NodeLogicalOp struct {
	Left  Node
	Right Node
	Op    LogicalOp
}

func (n NodeLogicalOp) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Left.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	b, ok := v.ToBool()
	if !ok {
		return values.Nil, values.FmtTypeError(n.Op.String(), values.KindBool)
	}

	if !b {
		return values.Of(false), nil
	}

	v, err = n.Right.Eval(e)
	if err != nil {
		return values.Nil, err
	}
	b, ok = v.ToBool()
	if !ok {
		return values.Nil, values.FmtTypeError(n.Op.String(), values.KindBool)
	}

	return values.Of(b), nil
}

func (n NodeLogicalOp) String() string {
	return n.Left.String() + " " + n.Op.String() + " " + n.Right.String()
}

type NodeForInLoop struct {
	Var      string
	Iterable Node
	Body     Node
}

func (n NodeForInLoop) Eval(e *Environment) (values.Value, *values.Error) {
	v, err := n.Iterable.Eval(e)
	if err != nil {
		return values.Nil, err
	}

	l, ok := v.ToList()
	if !ok {
		return values.Nil, values.FmtTypeError("for "+n.Var+" in ?", values.KindList)
	}

	for _, v := range l.AsSlice() {
		scope := e.NewChild()
		scope.SetScoped(n.Var, v)
		if _, err = n.Body.Eval(scope); err != nil {
			return values.Nil, err
		}
	}
	return values.Nil, nil
}

func (n NodeForInLoop) String() string {
	return "for " + n.Var + " in " + n.Iterable.String() + " " + n.Body.String()
}

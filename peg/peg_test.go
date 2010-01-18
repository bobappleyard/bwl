package peg

import (
	"strings";
	"testing";
)

func testExpr(e Expr, s string, srs bool, t *testing.T) {
	m := start("test", strings.NewReader(s));
	n, _ := e.Match(m);
	if n.Failed() == srs {
		t.Fail();
	}
}

func TestChar(t *testing.T) {
	testExpr(Char('h'), "hello", true, t);
	testExpr(Char('e'), "hello", false, t);
}

func TestString(t *testing.T) {
	testExpr(String("he"), "hello", true, t);
	testExpr(String("llo"), "hello", false, t);
}

func TestAnd(t *testing.T) {
	testExpr(And { String("he"), String("llo") },"hello", true, t);
	testExpr(And { String("llo"), String("he") }, "hello", false, t);
}

func TestOr(t *testing.T) {
	testExpr(Or { String("he"), String("llo") },"hello", true, t);
	testExpr(Or { String("llo"), String("he") }, "hello", true, t);
	testExpr(Or { String("llo"), String("lhe") }, "hello", false, t);
}

func TestModify(t *testing.T) {
	test := func(e Expr, s string, expected int) {
		m := start("test", strings.NewReader(s));
		n, _ := e.Match(m);
		if expected == -1 {
			if !n.Failed() {
				t.Fail();
			}
		} else if n.Pos()[0] != expected {
			t.Fail();
		}
	};
	test(Option(String("he")), "hello", 2);
	test(Option(String("llo")), "hello", 0);
	c := Char('a');
	test(Repeat(c), "", 0);
	test(Repeat(c), "a", 1);
	test(Repeat(c), "aa", 2);
	test(Repeat(c), "aaa", 3);
	test(Multi(c), "", -1);
	test(Multi(c), "a", 1);
	test(Multi(c), "aa", 2);
	test(Multi(c), "aaa", 3);
	test(Modify(c, 1, 1), "", -1);
	test(Modify(c, 1, 1), "a", 1);
	test(Modify(c, 1, 2), "aa", 2);
	test(Modify(c, 1, 3), "aaa", 3);
	test(Modify(And { c, Prevent(Any) }, 1, 2), "aaa", -1);
	test(Modify(c, 2, 3), "a", -1);
}

//~ func TestBind(t *testing.T) {
	//~ test := func(e Expr, s string

//~ }



package cql

import "testing"

func TestChainHelpers(t *testing.T) {
	q := (&Query{}).WithClause(
		Or(
			And(
				Search("id", "id1"),
				And(
					Search("side", "borrowing"),
					Search("requester_symbol", "PEER1"),
				),
			),
			And(
				Search("side", "lending"),
				Search("supplier_symbol", "PEER1"),
			),
		),
	)

	want := "id = id1 and (side = borrowing and requester_symbol = PEER1) or (side = lending and supplier_symbol = PEER1)"
	if got := q.String(); got != want {
		t.Fatalf("unexpected query string:\nwant: %s\ngot:  %s", want, got)
	}
}

func TestChainHelpersGrouping(t *testing.T) {
	q := (&Query{}).WithClause(
		And(
			Search("id", "id1"),
			Or(
				And(
					Search("side", "borrowing"),
					Search("requester_symbol", "PEER1"),
				),
				And(
					Search("side", "lending"),
					Search("supplier_symbol", "PEER1"),
				),
			),
		),
	)

	want := "id = id1 and (side = borrowing and requester_symbol = PEER1 or (side = lending and supplier_symbol = PEER1))"
	if got := q.String(); got != want {
		t.Fatalf("unexpected query string:\nwant: %s\ngot:  %s", want, got)
	}
}

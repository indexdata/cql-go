package parser

import (
	"testing"
)

func TestQueries(t *testing.T) {
	var p Parser

	node, err := p.Parse("beta")
	if err != nil || node == nil {
		t.Errorf("expected ok")
	}

	_, err = p.Parse("")
	if err == nil {
		t.Errorf("expected error")
	}

}

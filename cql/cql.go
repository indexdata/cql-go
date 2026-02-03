// Contextual Query Language (CQL) syntax tree API.
package cql

import (
	"strings"
)

// CQL built-in symbolic and named relations.
type Relation string

const (
	EQ       Relation = "="
	NE       Relation = "<>"
	LT       Relation = "<"
	GT       Relation = ">"
	LE       Relation = "<="
	GE       Relation = ">="
	ADJ      Relation = "adj"
	ALL      Relation = "all"
	ANY      Relation = "any"
	SCR      Relation = "scr"
	ENCLOSES Relation = "encloses"
	EXACT    Relation = "exact"
	WITHIN   Relation = "within"
)

// CQL boolean operators.
type Operator string

const (
	AND  Operator = "and"
	NOT  Operator = "not"
	OR   Operator = "or"
	PROX Operator = "prox"
)

// CQL special built-in indexes.
type CqlIndex string

const (
	AllRecords   CqlIndex = "cql.allRecords"
	AllIndexes   CqlIndex = "cql.allIndexes"
	AnyIndexes   CqlIndex = "cql.anyIndexes"
	Anywhere     CqlIndex = "cql.anywhere"
	Keywords     CqlIndex = "cql.keywords"
	ServerChoice CqlIndex = "cql.serverChoice"
	ResultSetId  CqlIndex = "cql.resultSetId"
)

// CQL built-in modifiers.
type CqlModifier string

const (
	Stem           CqlModifier = "stem"           //relation/sort, apply stemming to terms
	Relevant       CqlModifier = "relevant"       //relation/sort, apply relevancy matching
	Phonetic       CqlModifier = "phonetic"       //relation/sort, match words sounding like term
	Fuzzy          CqlModifier = "fuzzy"          //relation/sort, fuzzy matching
	Partial        CqlModifier = "partial"        //relation/sort, used with `within`/`encloses`
	IgnoreCase     CqlModifier = "ignoreCase"     //relation/sort, match case-insensitive
	RespectCase    CqlModifier = "respectCase"    //relation/sort, match case-sensitive
	IgnoreAccents  CqlModifier = "ignoreAccents"  //relation/sort, match w/o diacritics
	RespectAccents CqlModifier = "respectAccents" //relation/sort, match with diacritics
	Locale         CqlModifier = "locale"         //relation/sort, set locale, e.g `locale=da_DK`
	Word           CqlModifier = "word"           //relation/sort, break term into words
	String         CqlModifier = "string"         //relation/sort, treat term as a string
	IsoDate        CqlModifier = "isoDate"        //relation/sort, term is an ISO date, e.g `2000-10-31T01:30:00.000-05:00`
	Number         CqlModifier = "number"         //relation/sort, term is a number
	Uri            CqlModifier = "uri"            //relation/sort, term is a URI
	Oid            CqlModifier = "oid"            //relation/sort, terms is a ISO Object Identifier (OID), e.g `1.2.840.10003.3.1`
	Masked         CqlModifier = "masked"         //relation, on by default, * - zero/more chars, ? - single char, ^ - word anchor, \ - escape char, e.g `dc.title adj "*fish food*``
	Unmasked       CqlModifier = "unmasked"       //relation, disable masking
	Substring      CqlModifier = "substring"      //relation, match range of chars, e.g `marc.008 =/substring="1:6" 920102`
	Regexp         CqlModifier = "regexp"         //relation, term is a regular expression
	Distance       CqlModifier = "distance"       //proximity, distance expression e.g `cat prox/distance>2 hat`
	Unit           CqlModifier = "unit"           //proximity, unit for distance, one of paragraph, sentence, word (default), element
	Unordered      CqlModifier = "unordered"      //proximity, order of terms is unimportant
	Ordered        CqlModifier = "ordered"        //proximity, order of terms is important
)

// Represents the top-level CQL query.
// The `Clause“ field should be provided upon initialization.
// the `SortSpec` field is optional.
type Query struct {
	Clause
	SortSpec []Sort
}

func (q *Query) write(sb *strings.Builder) {
	q.Clause.write(sb, false)
	if len(q.SortSpec) > 0 {
		sb.WriteString(" sortBy")
	}
	for _, sort := range q.SortSpec {
		sb.WriteString(" ")
		sort.write(sb)
	}
}

func (q *Query) String() string {
	var sb strings.Builder
	q.write(&sb)
	return sb.String()
}

// Represents a sort criterion.
// If the `Index“ field is not set, the struct will stringify to an empty quoted string.
type Sort struct {
	Index     string
	Modifiers []Modifier
}

func (s *Sort) write(sb *strings.Builder) {
	quote(sb, s.Index)
	for _, mod := range s.Modifiers {
		sb.WriteString("/")
		mod.write(sb)
	}
}

func (s *Sort) String() string {
	var sb strings.Builder
	s.write(&sb)
	return sb.String()
}

// Represents a relation or a boolean operator modifier.
// At least `Name` should be set otherwise the struct will stringify to an empty quoted string.
// If `Relation` is not set but `Value` is, the `=` will be used during stringification.
type Modifier struct {
	Name     string
	Relation Relation
	Value    string
}

func (m *Modifier) write(sb *strings.Builder) {
	quote(sb, m.Name)
	if m.Value != "" {
		sb.WriteString(defVal(string(m.Relation), string(EQ)))
		quote(sb, m.Value)
	}
}

func (m *Modifier) String() string {
	var sb strings.Builder
	m.write(&sb)
	return sb.String()
}

// Represents either a search or a boolean clause (expression).
// Exactly one pointer field should be set when creating this struct.
// When neither pointer is set, the clause will stringify to `cql.allRecords = 1`
// (a special expression matching all records).
type Clause struct {
	PrefixMap    []Prefix
	SearchClause *SearchClause
	BoolClause   *BoolClause
}

func (c *Clause) write(sb *strings.Builder, brackets bool) {
	if len(c.PrefixMap) > 0 && brackets {
		sb.WriteString("(")
	}
	for _, p := range c.PrefixMap {
		p.write(sb)
		sb.WriteString(" ")
	}
	if c.SearchClause != nil {
		c.SearchClause.write(sb)
		if len(c.PrefixMap) > 0 && brackets {
			sb.WriteString(")")
		}
		return
	}
	if c.BoolClause != nil {
		if len(c.PrefixMap) == 0 && brackets {
			sb.WriteString("(")
		}
		c.BoolClause.write(sb)
		if brackets {
			sb.WriteString(")")
		}
		return
	}
	sb.WriteString(string(AllRecords))
	sb.WriteString(" ")
	sb.WriteString(string(EQ))
	sb.WriteString(" ")
	sb.WriteString("1")
	if len(c.PrefixMap) > 0 && brackets {
		sb.WriteString(")")
	}
}

func (c *Clause) String() string {
	var sb strings.Builder
	c.write(&sb, false)
	return sb.String()
}

// Represents a prefix declaration.
// The Prefix field can be omitted.
// The Uri field should be set or the struct will stringify to an empty quoted string.
type Prefix struct {
	Prefix string
	Uri    string
}

func (p *Prefix) write(sb *strings.Builder) {
	sb.WriteString("> ")
	if len(p.Prefix) > 0 {
		quote(sb, p.Prefix)
		sb.WriteString(" = ")
	}
	quote(sb, p.Uri)
}

func (p *Prefix) String() string {
	var sb strings.Builder
	p.write(&sb)
	return sb.String()
}

// Represents a search clause (expression).
// When no `Index` is given the `cql.serverChoice` is used during stringification.
// when no `Relation` is given the `=` (aka `scr`) is used during stringification.
type SearchClause struct {
	Index     string
	Relation  Relation
	Modifiers []Modifier
	Term      string
}

func (sc *SearchClause) write(sb *strings.Builder) {
	idx := defVal(sc.Index, string(ServerChoice))
	rel := defVal(string(sc.Relation), string(EQ))
	if idx != string(ServerChoice) ||
		(rel != string(EQ) && rel != string(SCR)) {
		quote(sb, idx)
		sb.WriteString(" ")
		sb.WriteString(rel)
		for _, mod := range sc.Modifiers {
			sb.WriteString("/")
			mod.write(sb)
		}
		sb.WriteString(" ")
	}
	quote(sb, sc.Term)
}

func (sc *SearchClause) String() string {
	var sb strings.Builder
	sc.write(&sb)
	return sb.String()
}

// Represents a boolean clause (expression).
// The `Left`, `Right` and `Operator` fields should be initialized.
// When the operator is not set, `and` will be used during stringification.
type BoolClause struct {
	Left      Clause
	Operator  Operator
	Modifiers []Modifier
	Right     Clause
}

func (bc *BoolClause) write(sb *strings.Builder) {
	bc.Left.write(sb, false)
	sb.WriteString(" ")
	sb.WriteString(defVal(string(bc.Operator), string(AND)))
	sb.WriteString(" ")
	bc.Right.write(sb, true)
}

func (bc *BoolClause) String() string {
	var sb strings.Builder
	bc.write(&sb)
	return sb.String()
}

const opsAndWhitespace = "()/<>= \t\r\n"

func quote(sb *strings.Builder, s string) {
	quote := s == ""
	escaped := false
	var term strings.Builder

	for _, ch := range s {
		if strings.ContainsRune(opsAndWhitespace, ch) {
			quote = true
		}
		if ch == '"' && !escaped {
			term.WriteByte('\\')
		}
		escaped = ch == '\\' && !escaped
		term.WriteRune(ch)
	}
	if escaped {
		// Trailing backslash: escape it.
		term.WriteByte('\\')
	}

	if quote {
		sb.WriteByte('"')
		sb.WriteString(term.String())
		sb.WriteByte('"')
	} else {
		sb.WriteString(term.String())
	}
}

func defVal(val string, def string) string {
	if len(val) > 0 {
		return val
	} else {
		return def
	}
}

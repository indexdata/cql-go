package pgcql

import (
	"fmt"
	"testing"
	"time"

	"github.com/indexdata/cql-go/cql"
	"github.com/stretchr/testify/require"
)

func TestDateTimeWithin(t *testing.T) {
	for _, tc := range []struct {
		name, lower, upper string
		onlyDate           bool
		upperRelation      string
		upperArgument      string
	}{
		{"dates", "2026-03-05", "2026-03-06", true, "<=", "2026-03-06"},
		{"datetime dates", "2026-03-05", "2026-03-06", false, "<", "2026-03-07"},
		{"timestamps", "2026-03-05T09:34:27Z", "2026-03-06T10:00:00+01:00", false, "<=", "2026-03-06T10:00:00+01:00"},
		{"space separated timestamps", "2026-03-05 09:34:27", "2026-03-06 10:00:00", false, "<=", "2026-03-06 10:00:00"},
		{"date and timestamp", "2026-03-05", "2026-03-06 10:00:00", false, "<=", "2026-03-06 10:00:00"},
		{"timestamp and date", "2026-03-05 09:34:27", "2026-03-06", false, "<", "2026-03-07"},
		{"equal endpoints", "2026-03-05", "2026-03-05", true, "<=", "2026-03-05"},
		{"reversed endpoints", "2026-03-06", "2026-03-05", true, "<=", "2026-03-05"},
		{"equal datetime dates", "2026-03-05", "2026-03-05", false, "<", "2026-03-06"},
		{"reversed datetime dates", "2026-03-06", "2026-03-05", false, "<", "2026-03-06"},
		{"month end", "2025-02-01", "2025-02-28", false, "<", "2025-03-01"},
		{"leap day", "2024-02-01", "2024-02-29", false, "<", "2024-03-01"},
		{"year end", "2025-12-01", "2025-12-31", false, "<", "2026-01-01"},
		{"explicit midnight", "2025-02-01", "2025-02-28T00:00:00Z", false, "<=", "2025-02-28T00:00:00Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			field := NewFieldDate().WithColumn("created_at")
			if tc.onlyDate {
				field.WithOnlyDate()
			}
			def := NewPgDefinition().AddField("date", field)
			var parser cql.Parser
			q, err := parser.Parse(fmt.Sprintf(`date within "%s %s"`, tc.lower, tc.upper))
			require.NoError(t, err)
			got, err := def.Parse(q, 3)
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("(created_at >= $3 AND created_at %s $4)", tc.upperRelation), got.GetWhereClause())
			q, err = parser.Parse(fmt.Sprintf(`date >= "%s" and date %s "%s"`, tc.lower, tc.upperRelation, tc.upperArgument))
			require.NoError(t, err)
			want, err := def.Parse(q, 3)
			require.NoError(t, err)
			require.Equal(t, want.GetQueryArguments(), got.GetQueryArguments())
		})
	}
}

func TestDateTimeWithinInvalid(t *testing.T) {
	for _, onlyDate := range []bool{false, true} {
		field := NewFieldDate().WithColumn("date")
		if onlyDate {
			field.WithOnlyDate()
		}
		for _, term := range []string{"", "2026-03-05", "invalid 2026-03-05", "2026-03-05 invalid", "2026-02-30 2026-03-05", "2026-03-05 2026-03-06 2026-03-07"} {
			_, args, err := field.Generate(cql.SearchClause{Relation: cql.WITHIN, Term: term}, 1)
			require.Error(t, err, "term %q, onlyDate %v", term, onlyDate)
			require.IsType(t, &PgError{}, err)
			if onlyDate {
				require.EqualError(t, err, fmt.Sprintf("invalid within range %q, expected two valid date endpoints in format YYYY-MM-DD", term))
			} else {
				require.EqualError(t, err, fmt.Sprintf("invalid within range %q, expected two valid date or date time endpoints", term))
			}
			require.Nil(t, args)
		}
	}
	_, _, err := NewFieldDate().WithOnlyDate().Generate(cql.SearchClause{Relation: cql.WITHIN, Term: "2026-03-05T09:34:27Z 2026-03-06"}, 1)
	require.Error(t, err)
}

func TestDateTimeWithinCombined(t *testing.T) {
	def := NewPgDefinition().AddField("date", NewFieldDate())
	var parser cql.Parser
	q, err := parser.Parse(`date = 2026-03-04 not date within "2026-03-05 2026-03-06" or date = 2026-03-07`)
	require.NoError(t, err)
	got, err := def.Parse(q, 2)
	require.NoError(t, err)
	require.Equal(t, "(date = $2 AND NOT (date >= $3 AND date < $4)) OR date = $5", got.GetWhereClause())
	require.Equal(t, []any{
		time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
	}, got.GetQueryArguments())
}

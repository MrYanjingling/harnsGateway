package audit

import "testing"

func TestQueryNormalize(t *testing.T) {
	q := Query{App: "  ems  "}
	q.Normalize()
	if q.App != AppEMS {
		t.Fatalf("app = %q, want %q", q.App, AppEMS)
	}
	if q.Page != 1 || q.PageSize != 10 {
		t.Fatalf("page/pageSize = %d/%d, want 1/10", q.Page, q.PageSize)
	}
}

func TestReaderScope(t *testing.T) {
	r := &reader{app: AppEMS}
	q := r.scope(Query{App: AppEAM})
	q.Page, q.PageSize = 2, 20
	if q.App != AppEMS {
		t.Fatalf("scoped app = %q, want %q", q.App, AppEMS)
	}

	all := (&reader{}).scope(Query{App: "ems"})
	if all.App != AppEMS {
		t.Fatalf("normalized app = %q, want %q", all.App, AppEMS)
	}
}

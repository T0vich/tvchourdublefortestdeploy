package main

import "testing"

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want []string
	}{
		{
			name: "пустой ввод",
			sql:  "   \n\t ",
			want: nil,
		},
		{
			name: "два простых оператора",
			sql:  "CREATE TABLE a(); CREATE TABLE b();",
			want: []string{"CREATE TABLE a()", "CREATE TABLE b()"},
		},
		{
			name: "хвост без завершающей точки с запятой",
			sql:  "SELECT 1; SELECT 2",
			want: []string{"SELECT 1", "SELECT 2"},
		},
		{
			name: "точка с запятой внутри строкового литерала",
			sql:  "INSERT INTO t VALUES ('a;b'); SELECT 1;",
			want: []string{"INSERT INTO t VALUES ('a;b')", "SELECT 1"},
		},
		{
			name: "строчный комментарий с кириллицей не ломает разбор",
			sql:  "-- Таблица пользователей; с точкой с запятой\nSELECT 1;",
			want: []string{"SELECT 1"},
		},
		{
			name: "блочный комментарий вырезается",
			sql:  "SELECT /* ; мусор ; */ 1; SELECT 2;",
			want: []string{"SELECT   1", "SELECT 2"},
		},
		{
			name: "блок DO с долларовыми кавычками остаётся целым",
			sql:  "TRUNCATE t CASCADE;\nDO $$ BEGIN INSERT INTO t VALUES (1); INSERT INTO t VALUES (2); END $$;",
			want: []string{
				"TRUNCATE t CASCADE",
				"DO $$ BEGIN INSERT INTO t VALUES (1); INSERT INTO t VALUES (2); END $$",
			},
		},
		{
			name: "именованная долларовая кавычка",
			sql:  "DO $body$ SELECT ';'; $body$; SELECT 1;",
			want: []string{"DO $body$ SELECT ';'; $body$", "SELECT 1"},
		},
		{
			name: "одиночный доллар не считается кавычкой",
			sql:  "SELECT 100$; SELECT 2;",
			want: []string{"SELECT 100$", "SELECT 2"},
		},
		{
			name: "незакрытый блочный комментарий не зацикливается",
			sql:  "SELECT 1; /* хвост без закрытия",
			want: []string{"SELECT 1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitStatements(tt.sql)

			if len(got) != len(tt.want) {
				t.Fatalf("получено %d операторов %q, ожидалось %d %q", len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("оператор %d: получено %q, ожидалось %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDollarTag(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"$$ BEGIN", "$$"},
		{"$body$ SELECT", "$body$"},
		{"$tag_1$x", "$tag_1$"},
		{"$ ", ""},
		{"$100", ""},
		{"$", ""},
	}

	for _, tt := range tests {
		if got := dollarTag(tt.in); got != tt.want {
			t.Errorf("dollarTag(%q) = %q, ожидалось %q", tt.in, got, tt.want)
		}
	}
}

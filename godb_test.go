package godb

import "testing"

func checkToSQL(t *testing.T, sqlExpected string, sqlProduced string, err error) {
	if err != nil {
		t.Fatal("ToSQL produces error :", err)
	}

	t.Log("SQL expected :", sqlExpected)
	t.Log("SQL produced :", sqlProduced)
	if sqlProduced != sqlExpected {
		t.Fatal("ToSQL produces incorrect SQL")
	}
}

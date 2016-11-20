package godb

// insertStatement is an INSERT statement builder.
type insertStatement struct {
	db *DB

	columns   []string
	intoTable string
	values    [][]interface{}
	suffixes  []string
}

// InsertInto initialise a INSERT INTO statement builder
func (db *DB) InsertInto(tableName string) *insertStatement {
	ip := &insertStatement{db: db}
	ip.intoTable = tableName
	return ip
}

// Columns add columns to insert.
func (ip *insertStatement) Columns(columns ...string) *insertStatement {
	ip.columns = append(ip.columns, columns...)
	return ip
}

// Values to insert
func (ip *insertStatement) Values(values ...interface{}) *insertStatement {
	ip.values = append(ip.values, values)
	return ip
}

// Suffix add an expression to suffix the statement.
func (ip *insertStatement) Suffix(suffix string) *insertStatement {
	ip.suffixes = append(ip.suffixes, suffix)
	return ip
}

// ToSQL returns a string with the SQL statement (containing placeholders),
// the arguments slices, and an error.
func (ip *insertStatement) ToSQL() (string, []interface{}, error) {

	// TODO : estimate the buffer size !!!
	sqlBuffer := newSQLBuffer(256, 16)

	sqlBuffer.write("INSERT ")

	if err := sqlBuffer.writeInto(ip.intoTable); err != nil {
		return "", nil, err
	}

	sqlBuffer.write(" (")
	if err := sqlBuffer.writeColumns(ip.columns); err != nil {
		return "", nil, err
	}
	sqlBuffer.write(") VALUES ")

	if err := sqlBuffer.writeInsertValues(ip.values, len(ip.columns)); err != nil {
		return "", nil, err
	}

	if err := sqlBuffer.writeStrings(ip.suffixes); err != nil {
		return "", nil, err
	}

	return sqlBuffer.sqlString(), sqlBuffer.sqlArguments(), nil
}

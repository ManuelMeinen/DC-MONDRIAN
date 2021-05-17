package types

import "database/sql"

const NullInt = 0
const NullString = ""

// return the value of i or 0 if i is NULL
func GetInt(i sql.NullInt32) (int){
	if i.Valid {
		return int(i.Int32)
	}else{
		return NullInt
	}
}

// return the value of s or "" if s is NULL
func GetString(s sql.NullString)(string){
	if s.Valid{
		return s.String
	}else{
		return NullString
	}
}
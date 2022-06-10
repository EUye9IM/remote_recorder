package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func toMD5(in string) string {
	data := md5.Sum([]byte(in))
	return hex.EncodeToString(data[:])
}

func dbinit() {
	var err error
	if config.Database.Initial {
		os.Remove(config.Database.Path)
	}
	db, err = sql.Open("sqlite3", config.Database.Path)
	if err != nil {
		log.Panicln("Open databases FAILED:", err)
	}
	if config.Database.Initial {
		_, err := db.Exec(`
			create table student (
				stu_no char(8) not null,
				stu_name char(16) not null,
				stu_password char(32) not null,
				stu_userlevel char(1) not null default '0',
				stu_enable char(1) not null default '1',
				primary key(stu_no)
			);`)
		if err != nil {
			log.Panicln("Initial databases ERROR:", err)
		}
		for _, v := range config.Database.Account {
			_, err := db.Exec(
				"insert into student values(?, ?, ?, ?, ?);",
				v.No, v.Name, toMD5(v.Password), v.Level, v.Enable)
			if err != nil {
				log.Panicln("Inseart data ERROR:", err)
			}
		}
	}
}
func login(user_no, password string) (Uinfo, bool) {
	uinfo := Uinfo{}
	row := db.QueryRow(`
		select stu_no, stu_name, stu_userlevel
		from student
		WHERE stu_no = ? and stu_password = ? and stu_enable = '1'`, user_no, toMD5(password))
	err := row.Scan(&uinfo.No, &uinfo.Name, &uinfo.Level)
	res := true
	if err != nil {
		res = false
	}
	return uinfo, res
}
func chpw(user_no, password string) bool {
	res, err := db.Exec(`
		update student
		set stu_password = ?
		where stu_no = ?`, toMD5(password), user_no)
	if err != nil {
		log.Panicln("Change password ERROR:", err)
	}
	num, err := res.RowsAffected()
	if err != nil {
		log.Panicln("Change password ERROR:", err)
	}
	if num == int64(0) {
		return false
	}
	return true
}

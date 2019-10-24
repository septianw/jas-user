package user

import (
	"fmt"
	"log"

	"database/sql"
	"errors"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	contact "github.com/septianw/jas-contact"
	"github.com/septianw/jas/common"
	"golang.org/x/crypto/bcrypt"
)

func getdbobj() (db *sql.DB, err error) {
	rt := common.ReadRuntime()
	dbs := common.LoadDatabase(filepath.Join(rt.Libloc, "database.so"), rt.Dbconf)
	db, err = dbs.OpenDb(rt.Dbconf)
	return
}

func Query(q string) (*sql.Rows, error) {
	db, err := getdbobj()
	common.ErrHandler(err)
	defer db.Close()

	return db.Query(q)
}

func Exec(q string) (sql.Result, error) {
	db, err := getdbobj()
	common.ErrHandler(err)
	defer db.Close()

	return db.Exec(q)
}

/*
CREATE TABLE IF NOT EXISTS `user` (
  `uid` INT NOT NULL AUTO_INCREMENT,
  `uname` VARCHAR(225) NOT NULL,
  `upass` TEXT NOT NULL,
  `contact_contactid` INT NOT NULL,
  PRIMARY KEY (`uid`,`uname`))
ENGINE = InnoDB;
*/

type User struct {
	Uid               int64
	Uname             string
	Upass             string
	Contact_contactid int64
}

type ContactType struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Prefix    string `json:"prefix"`
	Type      string `json:"type"`
}

type UserOut struct {
	Uid   int64  `json:"uid" binding:"required"`
	Uname string `json:"username" binding:"required"`
	ContactType
}

type UserIn struct {
	Uname string `json:"username" binding:"required"`
	ContactType
}

type UserAdd struct {
	Uname string `json:"username" binding:"required"`
	Upass string `json:"password" binding:"required"`
	ContactType
}

type UserUpdate struct {
	Upass string `json:"password" binding:"required"`
	ContactType
}

/*
ERROR CODE LEGEND:
error containt 4 digits,
first digit represent error location either module or main app
1 for main app
2 for module

second digit represent error at level app or database
1 for app
2 for database

third digit represent error with input variable or variable manipulation
0 for skipping this error
1 for input validation error
2 for variable manipulation error

fourth digit represent error with logic, this type of error have
increasing error number based on which part of code that error.
0 for skipping this error
1 for unknown logical error
2 for whole operation fail, operation end unexpectedly
*/

const DATABASE_EXEC_FAIL = 2200
const MODULE_OPERATION_FAIL = 2102
const INPUT_VALIDATION_FAIL = 2110

var NOT_ACCEPTABLE = gin.H{"code": "NOT_ACCEPTABLE", "message": "You are trying to request something not acceptible here."}
var NOT_FOUND = gin.H{"code": "NOT_FOUND", "message": "You are find something we can't found it here."}

// ini tipe-nya upsert.
// FIXME: still fail
func SetUser(userin UserAdd) (lastid int64, err error) {
	// var user UserOut
	// var kontak contact.ContactIn
	var user UserIn
	var userup UserUpdate

	user.Uname = userin.Uname

	users, err := FindUser(user)
	log.Printf("%+v", users)
	log.Printf("%+v", user)
	log.Printf("%+v", userin)
	if err != nil {
		return 0, err
	}

	if len(users) < 1 {
		return 0, errors.New("User not found.")
	}

	if len(users) < 1 {
		return InsertUser(userin)
	} else {
		userup.ContactType = userin.ContactType
		userup.Upass = userin.Upass
		return UpdateUser(users[0].Uid, userup)
	}

	return
}

func InsertUser(userin UserAdd) (lastid int64, err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var ctk contact.ContactIn
	var contactId int64
	var q string

	ctk.Firstname = userin.Firstname
	ctk.Lastname = userin.Lastname
	ctk.Prefix = userin.Prefix
	ctk.Type = userin.Type

	contacts, err := contact.FindContact(ctk)
	if err != nil {
		return 0, err
	}

	if len(contacts) == 0 {
		contactId, err = contact.InsertContact(ctk)
		if err != nil {
			return 0, err
		}
	} else {
		contactId = contacts[0].Id
	}

	/*
	  `uid` INT NOT NULL AUTO_INCREMENT,
	  `uname` VARCHAR(191) NOT NULL,
	  `upass` TEXT NOT NULL,
	  `deleted` TINYINT(1) NOT NULL DEFAULT 0,
	  `contact_contactid` INT NOT NULL,
	*/

	if strings.Compare(userin.Upass, "") == 0 {
		return 0, errors.New("Password cannot be empty.")
	}

	encrypted, err := bcrypt.GenerateFromPassword([]byte(userin.Upass), 10)
	if err != nil {
		return 0, err
	}

	log.Println(userin.Upass)
	log.Println(string(encrypted))
	log.Println(bcrypt.CompareHashAndPassword(encrypted, []byte(userin.Upass)))

	q = fmt.Sprintf(`
INSERT INTO user
	(uname, upass, deleted, contact_contactid)
VALUE ('%s', '%s', 0, %d)`, userin.Uname, encrypted, contactId)
	log.Println(q)

	result, err := Exec(q)
	if err != nil {
		return 0, nil
	}

	lastid, err = result.LastInsertId()
	if err != nil {
		return 0, nil
	}

	return
}

func TapUser(id, limit, offset int64) (usrs []User, err error) {
	var usr User
	var sbUser strings.Builder

	sbUser.WriteString("SELECT uid, uname, upass, contact_contactid FROM user WHERE deleted = 0")

	if id == -1 {
		if limit == 0 {
			sbUser.WriteString(" LIMIT 10 OFFSET 0")
		} else {
			sbUser.WriteString(fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset))
		}
	} else {
		if limit == 0 {
			sbUser.WriteString(fmt.Sprintf(" AND uid = %d LIMIT 10 OFFSET 0", id))
		} else {
			sbUser.WriteString(fmt.Sprintf(" AND uid = %d LIMIT %d OFFSET %d", id, limit, offset))
		}
	}

	rows, err := Query(sbUser.String())
	if err != nil {
		return
	}

	for rows.Next() {
		rows.Scan(&usr.Uid, &usr.Uname, &usr.Upass, &usr.Contact_contactid)
		usrs = append(usrs, usr)
	}

	if len(usrs) == 0 {
		err = errors.New("User not found.")
		return
	}

	return
}

func VerifyUser(username, hashedPassword string) (verified bool, err error) {
	var sbUser strings.Builder
	var encPassword string
	verified = false

	_, err = sbUser.WriteString(fmt.Sprintf(`SELECT upass FROM `+"`user`"+
		` WHERE uname = '%s'`, username))
	if err != nil {
		return
	}

	rows, err := Query(sbUser.String())
	if err != nil {
		return
	}

	for rows.Next() {
		rows.Scan(&encPassword)
	}

	err = bcrypt.CompareHashAndPassword([]byte(encPassword), []byte(hashedPassword))
	if err != nil {
		return
	} else {
		verified = true
	}

	return
}

func UpdateUser(uid int64, userin UserUpdate) (raff int64, err error) {
	var sbUpdate strings.Builder
	var set bool = false
	var upfield []string
	// var ctk contact.ContactIn
	// var ctkid int64

	_, err = sbUpdate.WriteString(`
		update user as u
		join contact as c on u.contact_contactid = c.contactid
		set `)
	if err != nil {
		return 0, err
	}

	usrs, err := GetUser(uid, 0, 0)
	if err != nil {
		return 0, err
	}
	if len(usrs) < 1 {
		return 0, errors.New("User not found.")
	}

	u := usrs[0]

	/*
	  `uid` INT NOT NULL AUTO_INCREMENT,
	  `uname` VARCHAR(191) NOT NULL,
	  `upass` TEXT NOT NULL,
	  `deleted` TINYINT(1) NOT NULL DEFAULT 0,
	  `contact_contactid` INT NOT NULL,
	*/

	if strings.Compare(u.Firstname, userin.Firstname) != 0 {
		set = true
		upfield = append(upfield, fmt.Sprintf("c.fname = '%s'", userin.Firstname))
	}
	if strings.Compare(u.Lastname, userin.Lastname) != 0 {
		set = true
		upfield = append(upfield, fmt.Sprintf("c.lname = '%s'", userin.Lastname))
	}
	if strings.Compare(u.Prefix, userin.Prefix) != 0 {
		set = true
		upfield = append(upfield, fmt.Sprintf("c.prefix = '%s'", userin.Prefix))
	}
	if strings.Compare(userin.Upass, "") != 0 {
		set = true
		encrypted, err := bcrypt.GenerateFromPassword([]byte(userin.Upass), 10)
		if err != nil {
			return 0, err
		}
		upfield = append(upfield, fmt.Sprintf("u.upass = '%s'", string(encrypted)))
	}

	if set {
		_, err = sbUpdate.WriteString(strings.Join(upfield, ", "))
		if err != nil {
			return 0, err
		}
		_, err = sbUpdate.WriteString(fmt.Sprintf(" where uid = %d", u.Uid))
		if err != nil {
			return 0, err
		}
	}

	q := sbUpdate.String()

	result, err := Exec(q)
	if err != nil {
		return 0, err
	}
	raff, _ = result.RowsAffected()

	return
}

// Get user by id.
func GetUser(id, limit, offset int64) ([]UserOut, error) {
	var record UserOut
	var records []UserOut
	var err error

	q := `select
	u.uid id,
    u.uname username,
    c.fname firstname,
    c.lname lastname,
    c.prefix prefix,
    ct.name type
from user u
join contact c on c.contactid = u.contact_contactid
join contactwtype cwt on c.contactid = cwt.contact_contactid
join contacttype ct on ct.ctypeid = cwt.contacttype_ctypeid
%s`
	constr := `limit %d offset %d`
	where := `where u.deleted = 0 and c.deleted = 0`
	whereid := `where u.deleted = 0 and c.deleted = 0 and u.uid = %d`

	if id == -1 {
		if limit == 0 {
			constr = fmt.Sprintf(constr, 10, 0)
			q = fmt.Sprintf(q, "%s %s")
			q = fmt.Sprintf(q, where, constr)
		} else {
			constr = fmt.Sprintf(constr, limit, offset)
			q = fmt.Sprintf(q, "%s %s")
			q = fmt.Sprintf(q, where, constr)
		}
	} else {
		q = fmt.Sprintf(q, fmt.Sprintf(whereid, id))
	}
	log.Println(q)

	rows, err := Query(q)
	common.ErrHandler(err)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err := rows.Scan(&record.Uid, &record.Uname, &record.Firstname, &record.Lastname, &record.Prefix, &record.Type)
		common.ErrHandler(err)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, err
}

// Query user.
func FindUser(userin UserIn) (records []UserOut, err error) {
	var user UserOut
	var sbuild strings.Builder
	// var where string
	q := `select
	u.uid id
from user u
join contact c on c.contactid = u.contact_contactid
where u.deleted = 0 and c.deleted = 0 %s`

	if strings.Compare(userin.Firstname, "") != 0 {
		_, err = sbuild.WriteString(fmt.Sprintf(`and c.fname = '%s' `, userin.Firstname))
		if err != nil {
			return
		}
	}

	if strings.Compare(userin.Lastname, "") != 0 {
		_, err = sbuild.WriteString(fmt.Sprintf(`and c.lname = '%s' `, userin.Lastname))
		if err != nil {
			return
		}
	}

	if strings.Compare(userin.Prefix, "") != 0 {
		_, err = sbuild.WriteString(fmt.Sprintf(`and c.prefix = '%s' `, userin.Prefix))
		if err != nil {
			return
		}
	}

	if strings.Compare(userin.Uname, "") != 0 {
		_, err = sbuild.WriteString(fmt.Sprintf(`and u.uname = '%s' `, userin.Uname))
		if err != nil {
			return
		}
	}

	q = fmt.Sprintf(q, sbuild.String())

	log.Printf("\n%+v\n", q)
	rows, err := Query(q)
	log.Printf("\n%+v err: %+v\n", rows, err)
	if err != nil {
		return
	}
	for rows.Next() {
		rows.Scan(&user.Uid)
		// contacts = append(contacts, contact)
	}
	records, err = GetUser(user.Uid, 0, 0)
	return
}

// Hapus user, set deleted jadi 1.
func DelUser(uid int64) (user UserOut, err error) {
	// var err error
	users, err := GetUser(uid, 0, 0)
	if err != nil {
		return
	}
	if len(users) == 0 {
		err = errors.New("User not found.")
		return
	}

	q := `UPDATE user SET deleted=1 WHERE uid = %d`
	result, err := Exec(fmt.Sprintf(q, users[0].Uid))
	if err != nil {
		return
	}

	affected, err := result.RowsAffected()
	log.Println(affected)
	if err != nil {
		return
	}

	user = users[0]

	return
}

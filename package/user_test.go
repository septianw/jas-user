package user

import (
	// "log"
	"fmt"
	"testing"

	"os"
	"reflect"

	"github.com/septianw/jas/common"
	"github.com/septianw/jas/types"
)

var uid int64

func SetEnvironment() {
	var rt types.Runtime
	var Dbconf types.Dbconf

	Dbconf.Database = "ipoint"
	Dbconf.Host = "localhost"
	Dbconf.Pass = "dummypass"
	Dbconf.Port = 3306
	Dbconf.Type = "mysql"
	Dbconf.User = "asep"

	rt.Dbconf = Dbconf
	rt.Libloc = "/home/asep/gocode/src/github.com/septianw/jas/libs"

	common.WriteRuntime(rt)
}

func UnsetEnvironment() {
	os.Remove("/tmp/shinyRuntimeFile")
}

func TestInsertUser(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	var userin UserAdd
	var err error

	userin.Firstname = "wawawa"
	userin.Lastname = "lalala"
	userin.Prefix = "ga"
	userin.Type = "konsumen"
	userin.Uname = "walaga"
	userin.Upass = "ini password saya."

	uid, err = InsertUser(userin)

	t.Log(uid)
	t.Log(err)

	if (uid == 0) && (err != nil) {
		t.Fail()
	}
}

func TestGetUser(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	users, err := GetUser(uid, 0, 0)
	t.Log(err)
	t.Log(users)
	if err != nil {
		t.Fail()
	}

	users, err = GetUser(-1, 5, 0)
	t.Log(err)
	t.Log(users)
	if err != nil {
		t.Fail()
	}
}

func TestFindUser(t *testing.T) {
	var userfind, userin, uout UserIn

	SetEnvironment()
	defer UnsetEnvironment()

	userfind.Firstname = "wawawa"
	userfind.Lastname = "lalala"
	userfind.Uname = "walaga"

	userin.Uname = "walaga"
	userin.Firstname = "wawawa"
	userin.Lastname = "lalala"
	userin.Prefix = "ga"
	userin.Type = "konsumen"

	users, err := FindUser(userin)
	t.Log(err)
	t.Log(users)
	if err != nil {
		t.Fail()
	}
	if len(users) == 0 {
		t.Fail()
	}

	uout.Firstname = users[0].Firstname
	uout.Lastname = users[0].Lastname
	uout.Prefix = users[0].Prefix
	uout.Type = users[0].Type
	uout.Uname = users[0].Uname

	t.Log(uout, userin)
	t.Log(reflect.DeepEqual(uout, userin))

	if uout != userin {
		t.Fail()
	}
}

func TestUpdateUser(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	var userin UserUpdate
	var userfind UserIn

	userin.Firstname = "Maybelle"
	userin.Lastname = "Ozelia"
	userin.Prefix = "Sir"
	userin.Type = "wow"
	userin.Upass = "40476a02377063c0f615375cd3f4b467"

	userfind.Firstname = userin.Firstname
	userfind.Lastname = userin.Lastname
	userfind.Prefix = userin.Prefix
	userfind.Type = userin.Type

	raff, err := UpdateUser(uid, userin)
	t.Log(raff)
	t.Log(err)

	users, err := GetUser(uid, 0, 0)

	if (userin.Firstname != users[0].Firstname) &&
		(userin.Lastname != users[0].Lastname) &&
		(userin.Prefix != users[0].Prefix) &&
		(userin.Type == users[0].Type) {
		t.Fail()
	}

	if (raff == 0) && (err != nil) {
		t.Fail()
	}
}

func TestSetUser(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	var userin UserAdd

	userin.Firstname = "Maybelle"
	userin.Lastname = "manoria"
	userin.Prefix = "Mr"
	userin.Type = "wow"
	userin.Upass = "40476a02377063c0f615375cd3f4b467"

	id, err := SetUser(userin)
	t.Log(id)
	t.Log(err)

	if (id == 0) && (err != nil) {
		t.Fail()
	}
}

func TestDelUser(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	var deleted int64

	user, err := DelUser(uid)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	users, err := GetUser(user.Uid, 0, 0)
	if len(users) > 0 {
		t.Fail()
	}

	q := fmt.Sprintf(`select deleted from user where uid = %d`, user.Uid)
	rows, err := Query(q)
	for rows.Next() {
		rows.Scan(&deleted)
	}

	if deleted == 0 {
		t.Log(deleted)
		t.Fail()
	}
}

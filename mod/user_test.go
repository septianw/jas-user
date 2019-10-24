package main

import (
	"fmt"
	"log"
	"testing"

	"crypto/md5"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"os"
	"strings"

	us "github.com/septianw/jas-user"
	"github.com/septianw/jas/common"
	"github.com/septianw/jas/types"
	"github.com/stretchr/testify/assert"
)

type header map[string]string
type headers []header
type payload struct {
	Method string
	Url    string
	Body   io.Reader
}
type expectation struct {
	Code int
	Body string
}

type quest struct {
	pload  payload
	heads  headers
	expect expectation
}
type quests []quest

var LastPostID int64

func getArm() (*gin.Engine, *httptest.ResponseRecorder) {
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	Router(router)

	recorder := httptest.NewRecorder()
	return router, recorder
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func doTheTest(load payload, heads headers) *httptest.ResponseRecorder {
	var router, recorder = getArm()

	req, err := http.NewRequest(load.Method, load.Url, load.Body)
	log.Printf("%+v", req)
	handleErr(err)

	if len(heads) != 0 {
		for _, head := range heads {
			for key, value := range head {
				req.Header.Set(key, value)
			}
		}
	}
	router.ServeHTTP(recorder, req)

	return recorder
}

func SetupRouter() *gin.Engine {
	return gin.New()
}

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

func TestPostUser(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	var useradd us.UserAdd
	var userin us.UserIn

	useradd.Firstname = "Firstname for test"
	useradd.Lastname = "Lastname for test"
	useradd.Prefix = "Tn."
	useradd.Type = "Konsumen"
	useradd.Uname = "lynda.alien"
	hashofpass := md5.Sum([]byte("newpasswordforlynda"))
	useradd.Upass = string(hashofpass[:])

	userin.Firstname = useradd.Firstname
	userin.Lastname = useradd.Lastname
	userin.Prefix = useradd.Prefix
	userin.Type = useradd.Type
	userin.Uname = useradd.Uname

	useraddJson, err := json.Marshal(useradd)
	common.ErrHandler(err)

	NewUser := strings.NewReader(string(useraddJson))

	q := quest{
		payload{"POST", "/api/v1/user/", NewUser},
		headers{},
		expectation{201, "contact post"},
	}

	rec := doTheTest(q.pload, q.heads)
	t.Log(rec)

	users, err := us.FindUser(userin)
	if err != nil || len(users) == 0 {
		t.Log(err)
		t.Log(users)
		t.Fail()
		return
	}
	t.Logf("\n%+v\n", users)
	LastPostID = users[0].Uid
	cjson, err := json.Marshal(users[0])
	if err != nil {
		t.Fail()
		return
	}

	assert.Equal(t, q.expect.Code, rec.Code)
	assert.Equal(t, string(cjson)+"\n", rec.Body.String())
}

func TestGetUser(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()

	qs := quests{
		quest{
			pload:  payload{"GET", fmt.Sprintf("/api/v1/user/%d", LastPostID), nil},
			heads:  headers{},
			expect: expectation{200, "contact post"},
		},
		quest{
			pload:  payload{"GET", "/api/v1/user/all/0/2", nil},
			heads:  headers{},
			expect: expectation{200, "contact post"},
		},
	}

	for _, q := range qs {
		rec := doTheTest(q.pload, q.heads)
		t.Log(rec)
		if strings.Contains(q.pload.Url, "all") {

			users, err := us.GetUser(-1, 2, 0)
			if (len(users) == 0) || (err != nil) {
				t.Logf("GetUser error: %+v %+v", users, err)
				t.Fail()
			}

			usersJson, err := json.Marshal(users)
			if err != nil {
				t.Logf("GetUser error: %+v", err)
				t.Fail()
			}

			assert.Equal(t, q.expect.Code, rec.Code)
			assert.Equal(t, string(usersJson), strings.TrimSpace(rec.Body.String()))
		} else {

			users, err := us.GetUser(LastPostID, 0, 0)
			if (len(users) == 0) || (err != nil) {
				t.Logf("GetUser error: %+v %+v", users, err)
				t.Fail()
			}

			usersJson, err := json.Marshal(users[0])
			if err != nil {
				t.Logf("GetUser error: %+v", err)
				t.Fail()
			}

			assert.Equal(t, q.expect.Code, rec.Code)
			assert.Equal(t, string(usersJson), strings.TrimSpace(rec.Body.String()))
		}
	}
}

func TestUserPutPositive(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()
	var useradd us.UserUpdate

	useradd.Firstname = "Johnny"
	useradd.Lastname = "Papa"
	useradd.Prefix = "Tn."
	useradd.Type = "Konsumen"
	hashofpass := md5.Sum([]byte("newpasswordforjohnny"))
	useradd.Upass = string(hashofpass[:])

	useraddJson, err := json.Marshal(useradd)
	if err != nil {
		t.Logf("Marshaller fail: %+v", err)
	}

	UpdateUser := strings.NewReader(string(useraddJson))
	// contactUpdatedJSON, err := json.Marshal(cpac.ContactOut{
	// 	LastPostID,
	// 	"Pramitha",
	// 	"Utami",
	// 	"Mr",
	// 	"konsumen",
	// })
	// common.ErrHandler(err)

	q := quest{
		payload{"PUT", fmt.Sprintf("/api/v1/user/%d", LastPostID), UpdateUser},
		headers{},
		expectation{200, ""},
	}

	rec := doTheTest(q.pload, q.heads)
	users, err := us.GetUser(LastPostID, 0, 0)
	if (err != nil) || (len(users) == 0) {
		t.Logf("Err, fail get User: %+v", err)
		t.Fail()
	}
	usersJson, err := json.Marshal(users[0])

	assert.Equal(t, q.expect.Code, rec.Code)
	assert.Equal(t, string(usersJson), strings.TrimSpace(rec.Body.String()))
}

func TestUserDelPositive(t *testing.T) {
	SetEnvironment()
	defer UnsetEnvironment()

	users, err := us.GetUser(LastPostID, 0, 0)
	if err != nil {
		t.Logf("Get user fail: %+v", err)
		t.Fail()
	}

	contactUpdatedJSON, err := json.Marshal(users[0])
	if err != nil {
		t.Logf("Marshall error: %+v", err)
		t.Fail()
	}

	q := quest{
		payload{"DELETE", fmt.Sprintf("/api/v1/user/%d", LastPostID), nil},
		headers{},
		expectation{200, string(contactUpdatedJSON)},
	}

	rec := doTheTest(q.pload, q.heads)

	assert.Equal(t, q.expect.Code, rec.Code)
	assert.Equal(t, q.expect.Body, strings.TrimSpace(rec.Body.String()))
}

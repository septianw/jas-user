package main

import (
	"fmt"
	"net/http"

	"github.com/septianw/jas/common"
	// "github.com/septianw/jas/types"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"database/sql"

	"github.com/gin-gonic/gin"
	// c "github.com/septianw/jas-contact/package"
	us "github.com/septianw/jas-user/package"
)

type Contact struct {
	Contactid int64
	Fname     string
	Lname     string
	Prefix    string
}

type User struct {
	Uid               int64
	Uname             string
	Upass             string
	Contact_contactid int64
}

type UserOut struct {
	Uid       int64  `json:"uid" binding:"required"`
	Username  string `json:"username" binding:"required"`
	FirstName string `json:"firstname" binding:"required"`
	LastName  string `json:"lastname" binding:"required"`
	Prefix    string `json:"prefix" binding:"required"`
}

/*
  `uid` INT NOT NULL AUTO_INCREMENT,
  `uname` VARCHAR(225) NOT NULL,
  `upass` TEXT NOT NULL,
  `contact_contactid` INT NOT NULL,
*/

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
const VERSION = us.Version

var NOT_ACCEPTABLE = gin.H{"code": "NOT_ACCEPTABLE", "message": "You are trying to request something not acceptible here."}
var NOT_FOUND = gin.H{"code": "NOT_FOUND", "message": "You are find something we can't found it here."}

var segments []string

func Bootstrap() {
	fmt.Println("wow")
}

/*
POST   /user
GET    /user/(:uid)
GET    /user/all/(:offset)/(:limit)
-----
ini masuk ke terminal
GET    /user/login
	basic auth
	return token, refresh token
-----
PUT    /user/(:uid)
DELETE /user/(:uid)
*/

func Router(r *gin.Engine) {
	// db := common.LoadDatabase()
	r.Any("/api/v1/user/*path1", deflt)
	// r.GET("/user/list", func(c *gin.Context) {
	// 	c.String(http.StatusOK, "wow")
	// })
}

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

func deflt(c *gin.Context) {
	segments = strings.Split(c.Param("path1"), "/")
	// log.Printf("\n%+v\n", c.Request.Method)
	// log.Printf("\n%+v\n", c.Param("path1"))
	// log.Printf("\n%+v\n", segments)
	// log.Printf("\n%+v\n", len(segments))

	switch c.Request.Method {
	case "POST":
		if strings.Compare(segments[1], "") == 0 {
			// dummyResponse(c)
			PostUserHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, NOT_ACCEPTABLE)
		}
		break
	case "GET":
		if strings.Compare(segments[1], "all") == 0 {
			GetUserAllHandler(c)
		} else if i, e := strconv.Atoi(segments[1]); (e == nil) && (i > 0) {
			GetUserIdHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, NOT_ACCEPTABLE)
		}
		break
	case "PUT":
		if i, e := strconv.Atoi(segments[1]); (e == nil) && (i > 0) {
			PutUserHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, NOT_ACCEPTABLE)
		}
		break
	case "DELETE":
		if i, e := strconv.Atoi(segments[1]); (e == nil) && (i > 0) {
			DelUserHandler(c)
		} else {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, NOT_ACCEPTABLE)
		}
		break
	default:
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, NOT_ACCEPTABLE)
		break
	}
}

func dummyResponse(c *gin.Context) {
	c.String(http.StatusOK, "wow")
}

func PostUserHandler(c *gin.Context) {
	var input us.UserAdd

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": us.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
		return
	}

	uid, err := us.InsertUser(input)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": us.DATABASE_EXEC_FAIL,
			"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
		return
	}

	users, err := us.GetUser(uid, 0, 0)
	if err != nil {
		if strings.Compare("User not found.", err.Error()) == 0 {
			c.JSON(http.StatusNotFound, us.NOT_FOUND)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": us.DATABASE_EXEC_FAIL,
				"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
			return
		}
	}

	if len(users) == 0 {
		c.JSON(http.StatusNotFound, us.NOT_FOUND)
		return
	}

	c.JSON(http.StatusCreated, users[0])
}

func GetUserIdHandler(c *gin.Context) {
	var id int64 = 0

	i, e := strconv.Atoi(segments[1])

	if e != nil { // konversi berhasil
		c.JSON(http.StatusBadRequest, gin.H{"code": us.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", e.Error())})
		return
	}
	id = int64(i)

	records, err := us.GetUser(id, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": us.DATABASE_EXEC_FAIL,
			"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
		return
	}
	if len(records) == 0 {
		c.JSON(http.StatusNotFound, us.NOT_FOUND)
		return
	}
	record := records[0]

	c.JSON(http.StatusOK, record)
	return
}

func GetUserAllHandler(c *gin.Context) {
	var l, o int64
	var limit, offset int
	var err error

	if len(segments) == 3 {
		limit = 10
		offset, err = strconv.Atoi(segments[2])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": us.INPUT_VALIDATION_FAIL,
				"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
			return
		}
	} else if len(segments) == 4 {
		limit, err = strconv.Atoi(segments[3])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": us.INPUT_VALIDATION_FAIL,
				"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
			return
		}
		offset, err = strconv.Atoi(segments[2])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": us.INPUT_VALIDATION_FAIL,
				"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
			return
		}
	} else {
		limit = 10
		offset = 0
	}

	if err == nil { // tidak ada error dari konversi
		l = int64(limit)
		o = int64(offset)
	}

	records, err := us.GetUser(-1, l, o)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": us.DATABASE_EXEC_FAIL,
			"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, records)
	return
}

func PutUserHandler(c *gin.Context) {
	var id int64
	var input us.UserUpdate

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": us.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", err.Error())})
		return
	}

	i, e := strconv.Atoi(segments[1])
	if e == nil { // konversi berhasil
		id = int64(i)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": us.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", e.Error())})
		return
	}

	_, err := us.UpdateUser(id, input)

	if err != nil {
		if strings.Compare("User not found.", err.Error()) == 0 {
			c.JSON(http.StatusNotFound, us.NOT_FOUND)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": us.DATABASE_EXEC_FAIL,
				"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
			return
		}
	}

	on, err := us.GetUser(id, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": us.DATABASE_EXEC_FAIL,
			"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
		return
	}
	log.Println(on)

	c.JSON(http.StatusOK, on[0])
	return
}

func DelUserHandler(c *gin.Context) {
	var id int64 = 1

	i, e := strconv.Atoi(segments[1])
	if e == nil { // konversi berhasil
		id = int64(i)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": us.INPUT_VALIDATION_FAIL,
			"message": fmt.Sprintf("INPUT_VALIDATION_FAIL: %s", e.Error())})
		return
	}

	// contacts := us.GetContact(id, 0, 0)
	user, err := us.DelUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": us.DATABASE_EXEC_FAIL,
			"message": fmt.Sprintf("DATABASE_EXEC_FAIL: %s", err.Error())})
		return
	} else if (err != nil) && (strings.Compare("User not found.", err.Error()) == 0) {
		c.JSON(http.StatusNotFound, us.NOT_FOUND)
		return
	}

	c.JSON(http.StatusOK, user)
}

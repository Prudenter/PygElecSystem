package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

//定义用户类
type User struct {
	Id int
	Name string 			`orm:"size(40);unique"`
	Pwd string 				`orm:"size(40)"`
	Phone string 			`orm:"size(11)"`
	Email string 			`orm:"null"`
	Active bool 			`orm:"default(false)"`
	Addresses []*Address 	`orm:"reverse(many)"`
}

//定义地址类，并与用户类建立1：n关系
type Address struct {
	Id int
	Receiver string 		`orm:"size(40)"`
	Addr string 			`orm:"size(100)"`
	PostCode string 		`orm:"size(6)"`
	Phone string 			`orm:"size(11)"`
	IsDefault bool			`orm:"default(false)"`  //是否时默认地址
	User *User 				`orm:"rel(fk)"`
}

func init(){
	//注册数据库
	orm.RegisterDataBase("default","mysql","root:123456@tcp(127.0.0.1:3306)/pygEs")
	//注册表结构
	orm.RegisterModel(new(User),new(Address))
	//运行生成表
	orm.RunSyncdb("default",false,true)
}
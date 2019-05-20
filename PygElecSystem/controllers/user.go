package controllers

import (
	"github.com/astaxie/beego"
	"fmt"
	"regexp"
	"github.com/astaxie/beego/orm"
	"PygElecSystem/PygElecSystem/models"
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"time"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"math/rand"
	"github.com/astaxie/beego/utils"
	"strings"
)

/* 定义用户控制器类:UserController */

type UserController struct {
	beego.Controller
}

/* 定义消息类 */
type Message struct {
	Message   string
	RequestId string
	BizId     string
	Code      string
}

/* 定义函数负责用户注册页面展示 */
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

/* 定义函数负责发送短信验证码 */
func (this *UserController) HandleSendMsg() {
	//接受页面传来的数据
	phone := this.GetString("phone")
	//1.定义一个传递给ajax json数据的容器
	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller, resp)
	//返回json格式数据
	//校验数据
	if phone == "" {
		fmt.Println("获取电话号码失败！")
		//2.给容器赋值
		resp["errno"] = 1
		resp["errmsg"] = "获取电话号码失败!"
		return
	}
	//检查电话号码格式是否正确
	reg, _ := regexp.Compile(`^1[3-9][0-9]{9}$`)
	ret := reg.FindString(phone)
	if ret == "" {
		fmt.Println("电话号码格式错误！")
		//2.给容器赋值
		resp["errno"] = 2
		resp["errmsg"] = "电话号码格式错误!"
		return
	}

	//发送短信，调用阿里云SDK
	client, err := sdk.NewClientWithAccessKey("cn-hangzhou", "LTAIu4sh9mfgqjjr", "sTPSi0Ybj0oFyqDTjQyQNqdq9I9akE")
	if err != nil {
		fmt.Println("初始化短信错误！")
		//2.给容器赋值
		resp["errno"] = 3
		resp["errmsg"] = "初始化短信错误"
		return
	}

	//生成6位数随机数
	var authCode string
	//创建随机数种子
	//rand.Seed(time.Now().UnixNano())
	//for i := 0; i < 6; i++ {
	//	//生成1-100以内的随机数
	//	num := rand.Intn(10)
	//	authCode += strconv.Itoa(num)
	//}
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	authCode = fmt.Sprintf("%06d", rand.Int31n(1000000))

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-hangzhou"
	request.QueryParams["PhoneNumbers"] = phone
	request.QueryParams["SignName"] = "品优购"
	request.QueryParams["TemplateCode"] = "SMS_164275022"
	request.QueryParams["TemplateParam"] = "{\"code\":" + authCode + "}"
	//调用方法，发送随机验证码短信
	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		fmt.Println("短信发送失败！")
		//2.给容器赋值
		resp["errno"] = 4
		resp["errmsg"] = "短信发送失败"
		return
	}

	//json数据解析
	var message Message
	json.Unmarshal(response.GetHttpContentBytes(), &message)
	//json.Unmarshal([]byte(response.String()),&message)
	if message.Message != "OK" {
		fmt.Println("json数据解析失败！")
		//2.给容器赋值
		resp["errno"] = 6
		resp["errmsg"] = message.Message
		return
	}
	resp["errno"] = 5
	resp["errmsg"] = "发送成功！"
	resp["code"] = authCode
}

/* 定义函数，负责统一返回json数据到前台 */
func RespFunc(this *beego.Controller, resp map[string]interface{}) {
	//3.把容器传递给前端
	this.Data["json"] = resp
	//4.指定传递方式，以json格式传递数据
	this.ServeJSON()
}

/* 定义函数，负责用户注册业务处理 */
func (this *UserController) HandleRegister() {
	//获取数据
	phone := this.GetString("phone")
	pwd := this.GetString("password")
	repwd := this.GetString("repassword")
	//校验数据
	if phone == "" || pwd == "" || repwd == "" {
		fmt.Println("获取数据错误")
		this.Data["errmsg"] = "获取数据错误"
		this.TplName = "register.html"
		return
	}
	if pwd != repwd {
		fmt.Println("两次密码输入不一致!")
		this.Data["errmsg"] = "两次密码输入不一致"
		this.TplName = "register.html"
		return
	}
	//处理数据
	//orm插入数据
	o := orm.NewOrm()
	var user models.User
	user.Name = phone
	user.Pwd = pwd
	user.Phone = phone
	_, err := o.Insert(&user) //这里应该判断重复插入问题  errmsg需要在页面接收
	if err != nil {
		fmt.Println(err)
		this.Data["errmsg"] = "注册失败，请重新注册!"
		this.TplName = "register.html"
		return
	}
	//注册成功,设置cookie,用于邮箱激活
	this.Ctx.SetCookie("userName", user.Name, 60*10)
	//返回数据,跳转到邮箱激活页面
	this.Redirect("/register-email", 302)
}

/* 定义函数,负责邮箱激活页面展示 */
func (this *UserController) ShowRegisterEmail() {
	this.TplName = "register-email.html"
}

/* 定义函数,负责邮箱激活业务处理 */
func (this *UserController) HandleRegisterEmail() {
	//获取数据
	email := this.GetString("email")
	pwd := this.GetString("password")
	repwd := this.GetString("repassword")
	//校验数据
	if email == "" || pwd == "" || repwd == "" {
		fmt.Println("获取数据错误")
		this.Data["errmsg"] = "获取数据错误"
		this.TplName = "register-email.html"
		return
	}
	if pwd != repwd {
		fmt.Println("两次密码输入不一致!")
		this.Data["errmsg"] = "两次密码输入不一致"
		this.TplName = "register-email.html"
		return
	}
	//校验邮箱格式
	reg, _ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	ret := reg.FindString(email)
	if ret == "" {
		fmt.Println("邮箱格式错误")
		this.Data["errmsg"] = "邮箱格式错误"
		this.TplName = "register-email.html"
		return
	}
	//处理数据,发送邮件
	//定义发送邮件的配置
	config := `{"username":"18706733725@163.com","password":"qwer123","host":"smtp.163.com","port":25}`
	//定义邮件对象
	emailReg := utils.NewEMail(config)
	//配置邮件内容
	emailReg.Subject = "品优购用户激活链接"
	emailReg.From = "18706733725@163.com"
	emailReg.To = []string{email}
	//通过cookie获取用户名
	userName := this.Ctx.GetCookie("userName")
	emailReg.HTML = `<a href="http://127.0.0.1:8080/active?userName=` + userName + `">点击此处激活用户</a>`
	//发送邮件
	err := emailReg.Send()
	if err != nil {
		fmt.Println("邮箱发送失败")
		this.Data["errmsg"] = "邮箱发送失败"
		this.TplName = "register-email.html"
		return
	}
	//邮件发送成功后,插入邮件
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err = o.Read(&user, "Name")
	if err != nil {
		fmt.Println("注册失败,请重新注册")
		this.Data["errmsg"] = "注册失败,请重新注册"
		this.TplName = "register.html"
		return
	}
	//插入邮件字段
	user.Email = email
	//修改邮箱!!
	o.Update(&user)
	//返回数据
	this.Ctx.WriteString("邮件已发送，请去您的邮箱激活用户！")
}

/* 定义函数,负责用户激活业务处理 */
func (this *UserController) HandleActive() {
	//获取数据
	userName := this.GetString("userName")
	//校验数据
	if userName == "" {
		fmt.Println("用户名错误")
		this.Data["errmsg"] = "用户名错误"
		this.Redirect("/register-email", 302)
		return
	}
	//处理数据,更新active
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		fmt.Println("err :", err)
		fmt.Println()
		fmt.Println("用户名不存在")
		this.Data["errmsg"] = "用户名不存在"
		this.Redirect("/register-email", 302)
		return
	}
	user.Active = true
	o.Update(&user, "Active")
	//返回数据
	this.Redirect("/login", 302)
}

/* 定义函数,负责用户登录页面展示 */
func (this *UserController) ShowLogin() {
	//获取cookie,如果有值,说明已勾选记住用户名,则页面需要勾选记住用户名,否则不勾选
	userName := this.Ctx.GetCookie("userName")
	if userName != "" {
		this.Data["checked"] = "checked"
	} else {
		this.Data["checked"] = ""
	}
	this.Data["userName"] = userName
	//指定视图页面
	this.TplName = "login.html"
}

/* 定义函数,负责用户登录业务为处理 */
func (this *UserController) HandleLogin() {
	//获取数据
	userName := this.GetString("userName")
	password := this.GetString("password")
	m1, _ := this.GetInt("m1")
	//校验数据
	if userName == "" || password == "" {
		fmt.Println("用户名或密码不能为空")
		this.Data["errmsg"] = "用户名或密码不能为空"
		this.TplName = "login.html"
		return
	}
	//处理数据,查询校验
	o := orm.NewOrm()
	var user models.User
	//赋值
	reg, _ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	ret := reg.FindString(userName)
	if ret != "" {
		user.Email = userName
		err := o.Read(&user, "Email")
		if err != nil {
			this.Data["errmsg"] = "邮箱未注册!"
			this.TplName = "login.html"
			return
		}
		//校验密码
		if user.Pwd != password {
			this.Data["errmsg"] = "输入密码不正确!"
			this.TplName = "login.html"
			return
		}
	} else {
		user.Name = userName
		err := o.Read(&user, "Name")
		if err != nil {
			this.Data["errmsg"] = "用户名不存在!"
			this.TplName = "login.html"
			return
		}
		//校验密码
		if user.Pwd != password {
			this.Data["errmsg"] = "输入密码不正确!"
			this.TplName = "login.html"
			return
		}
	}

	//校验用户是否激活
	if user.Active == false {
		this.Data["errmsg"] = "当前用户未激活，请去目标邮箱激活！"
		this.TplName = "login.html"
		return
	}

	//根据m1的值,判断是否实现记住用户名
	if m1 == 2 {
		this.Ctx.SetCookie("userName", userName, 60*100)
	} else {
		//设置存活时间为-1,使cookie失效
		this.Ctx.SetCookie("userName", userName, -1)
	}
	//设置session,用于登陆后页面使用
	this.SetSession("userName", user.Name)
	//返回数据
	this.Redirect("/index", 302)
}

/* 定义函数,负责退出登录业务处理 */
func (this *UserController) Logout() {
	//删除session
	this.DelSession("userName")
	//跳转页面
	this.Redirect("/index", 302)
}

/* 定义函数,负责用户中心个人信息页面展示 */
func (this *UserController) ShowUserCenterInfo() {
	//调用函数,获取当前登录用户
	user := GetUser(&this.Controller)
	this.Data["user"] = user
	//调用函数,获取当前登录用户的默认地址
	this.Data["address"] = GetUserAddr(&this.Controller)
	//实现视图布局,将模板与主要部分连接其起来
	this.Layout = "user_center_layout.html"
	this.Data["num"] = 1
	this.TplName = "user_center_info.html"
}

/* 定义函数,获取当前登录用户 */
func GetUser(this *beego.Controller) models.User {
	//根据session获取当前登录用户名
	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user,"Name")
	//手机号码加密
	str := user.Phone
	user.Phone = strings.Join([]string{str[0:3],"****",str[7:]},"");
	return user
}

/* 定义函数,查询当前用户默认地址*/
func GetUserAddr(this *beego.Controller) models.Address {
	//查询数据库,显示默认地址
	o := orm.NewOrm()
	var address models.Address
	//获取当前用户的默认地址
	userName := this.GetSession("userName").(string)
	qs := o.QueryTable("Address")
	qs.RelatedSel("User").Filter("User__Name", userName).Filter("IsDefault", true).One(&address)
	//手机号码加密
	if address.Phone != "" {
		str := address.Phone
		address.Phone = strings.Join([]string{str[0:3], "****", str[7:]}, "")
	}
	return address
}


/* 定义函数,负责收货地址页面展示 */
func (this *UserController) ShowSite() {
	//调用函数,获取当前登录用户
	user := GetUser(&this.Controller)
	this.Data["user"] = user
	//调用函数,获取当前登录用户的默认地址
	this.Data["address"] = GetUserAddr(&this.Controller)
	//实现视图布局,将模板与主要部分连接其起来
	this.Layout = "user_center_layout.html"
	this.Data["num"] = 3
	this.TplName = "user_center_site.html"
}

/* 定义函数,负责新增地址业务处理 */
func (this *UserController) HandleSite() {
	//获取数据
	receiver := this.GetString("receiver")
	addr := this.GetString("addr")
	postCode := this.GetString("postCode")
	phone := this.GetString("phone")
	//校验数据
	if receiver == "" || addr == "" || postCode == "" || phone == "" {
		fmt.Println("收件人,收货地址或联系电话不能为空!")
		this.Data["errmsg"] = "收件人,收货地址或联系电话不能为空!"
		this.Redirect("/user/site", 302)
		return
	}
	//检查电话号码格式是否正确
	reg, _ := regexp.Compile(`^1[3-9][0-9]{9}$`)
	ret := reg.FindString(phone)
	if ret == "" {
		fmt.Println("电话号码格式错误！")
		this.Data["errmsg"] = "电话号码格式错误!"
		this.Redirect("/user/site", 302)
		return
	}

	//处理数据,插入数据
	o := orm.NewOrm()
	var address models.Address
	//调用函数,获取当前登录用户
	user := GetUser(&this.Controller)
	//给插入对象赋值
	address.Addr = addr
	address.User = &user
	address.Phone = phone
	address.PostCode = postCode
	address.Receiver = receiver
	//设置当前插入的address为默认地址-
	// 要先查询当前用户是否有默认地址,如果有,则把默认地址修改为非默认,在设置当前地址为默认地址;如果没有直接设置为默认地址
	var defaultAddr models.Address
	qs := o.QueryTable("Address")
	err := qs.RelatedSel("User").Filter("User__Name", user.Name).Filter("IsDefault", true).One(&defaultAddr)
	if err == nil {
		defaultAddr.IsDefault = false
		o.Update(&defaultAddr, "IsDefault")
	}
	address.IsDefault = true
	//插入当前地址
	_, err = o.Insert(&address)
	if err != nil {
		fmt.Println("地址插入失败！")
		this.Data["errmsg"] = "地址插入失败!"
		this.Redirect("/user/site", 302)
		return
	}
	//返回数据
	this.Redirect("/user/site", 302)
}

/* 定义函数,负责展示当前用户全部订单 */
func (this *UserController) ShowUserOrder() {
	//调用函数,获取当前登录用户
	user := GetUser(&this.Controller)
	this.Data["user"] = user
	////调用函数,获取当前登录用户的默认地址
	this.Data["address"] = GetUserAddr(&this.Controller)
	//实现视图布局,将模板与主要部分连接其起来
	this.Layout = "user_center_layout.html"
	//用于页面样式判断
	this.Data["num"] = 2
	this.TplName = "user_center_order.html"
}
package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"strings"
	"PygElecSystem/PygElecSystem/models"
)

type GoodsController struct {
	beego.Controller
}

/* 定义函数,负责商品首页展示 */
func (this *GoodsController) ShowIndex() {
	//获取session,用于页面显示
	userName := this.GetSession("userName")
	if userName != nil {
		this.Data["userName"] = userName.(string)
	} else {
		this.Data["userName"] = ""
	}
	//获取类型信息返回到前台
	//定义总容器,一个map切片,value为interface
	var typeSlice []map[string]interface{}
	//获取所有一级菜单
	o := orm.NewOrm()
	//定义切片,获取所有一级菜单
	var firstTypes []models.TpshopCategory
	o.QueryTable("TpshopCategory").Filter("Pid", 0).All(&firstTypes)

	//遍历所有一级菜单,获取每个一级菜单下的二级菜单'
	for _, firstType := range firstTypes {
		//定义行容器,将每个一级菜与其对应的子菜单绑定到一起
		firstRows := make(map[string]interface{})
		//定义切片,获取所有二级菜单
		var secondTypes []models.TpshopCategory
		o.QueryTable("TpshopCategory").Filter("Pid", firstType.Id).All(&secondTypes)
		//把一级菜单存入map
		firstRows["first"] = firstType

		//定义二级容器,存放所有二级行容器
		var sencondTypeSlice []map[string]interface{}

		//遍历所有二级菜单,获取每个二级菜单下的三级菜单
		for _, secondType := range secondTypes {
			//定义行容器,将每个二级菜单及其对应的三级菜单绑定到一行
			secondRows := make(map[string]interface{})
			//定义切片,存放所有三级级菜单
			var thirdTypes []models.TpshopCategory
			o.QueryTable("TpshopCategory").Filter("Pid", secondType.Id).All(&thirdTypes)
			secondRows["second"] = secondType
			secondRows["third"] = thirdTypes
			//把每个二级行容器存入二级切片容器里
			sencondTypeSlice = append(sencondTypeSlice, secondRows)
		}
		//把二级行容器存入一级行容器中,将每个二级行容器与其对应的一级菜单绑定
		firstRows["son"] = sencondTypeSlice
		//把一级行容器存入总容器中
		typeSlice = append(typeSlice, firstRows)
	}
	//返回数据到页面
	this.Data["typeSlice"] = typeSlice
	this.TplName = "index.html"
}

/* 定义函数,获取当前登录用户 */
func GoodsGetUser(this *GoodsController) models.User {
	//根据session获取当前登录用户名
	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")
	//手机号码加密
	str := user.Phone
	user.Phone = strings.Join([]string{str[0:3], "****", str[7:]}, "");
	return user
}

/* 定义函数,查询当前用户默认地址*/
func GoodsGetUserAddr(this *GoodsController) models.Address {
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

/* 定义函数,负责全部订单页面显示 */
func (this *GoodsController) ShowOrder() {
	//调用函数,获取当前登录用户
	user := GoodsGetUser(this)
	this.Data["user"] = user
	////调用函数,获取当前登录用户的默认地址
	this.Data["address"] = GoodsGetUserAddr(this)
	//实现视图布局,将模板与主要部分连接其起来
	this.Layout = "user_center_layout.html"
	this.Data["num"] = 2
	this.TplName = "user_center_order.html"
}

/* 定义函数,负责生鲜模块首页展示 */
func (this *GoodsController) ShowIndex_sx() {
	//获取生鲜首页内容

	//获取商品类型
	var goodsTypes []models.GoodsType
	o := orm.NewOrm()
	o.QueryTable("GoodsType").All(&goodsTypes)
	//返回所有商品类型到页面
	this.Data["goodsTypes"] = goodsTypes

	//获取首页所有轮播图
	var goodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsBanners)
	//返回首页所有轮播图到页面
	this.Data["goodsBanners"] = goodsBanners

	//获取所有促销商品
	var promotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotionBanners)
	//返回所有促销商品到页面
	this.Data["promotionBanners"] = promotionBanners

	//首页商品展示
	//定义纵容器,保存所有一级菜单及其子菜单
	var goods []map[string]interface{}
	//循环所有一级菜单,获取其子菜单
	for _,goodType := range goodsTypes{
		//定义行容器,将一级菜单与其对应的子菜单绑定到一起
		rows := make(map[string]interface{})
		rows["goodType"] = goodType
		//定义切片存取所有文字商品
		var textGoods []models.IndexTypeGoodsBanner
		qs := o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").
			Filter("GoodsType__Id",goodType.Id).OrderBy("Index")
		//获取所有文字商品
		qs.Filter("DisplayType",0).All(&textGoods)
		rows["textGoods"] = textGoods
		//定义切片存取所有图片商品
		var imageGoods []models.IndexTypeGoodsBanner
		//获取所有图片商品
		qs.Filter("DisplayType",1).All(&imageGoods)
		rows["imageGoods"] = imageGoods
		//将行容器存入总容器中
		goods = append(goods,rows)
	}
	//返回数据到前端
	this.Data["goods"] = goods
	//返回数据
	this.TplName = "index_sx.html"
}

/* 定义函数,负责商品详情页面展示 */
func (this *GoodsController) ShowGoodsDetail() {
	this.TplName = "detail.html"
}
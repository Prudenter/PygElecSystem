package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"PygElecSystem/PygElecSystem/models"
	"fmt"
	"math"
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
	for _, goodType := range goodsTypes {
		//定义行容器,将一级菜单与其对应的子菜单绑定到一起
		rows := make(map[string]interface{})
		rows["goodType"] = goodType
		//定义切片存取所有文字商品
		var textGoods []models.IndexTypeGoodsBanner
		qs := o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").
			Filter("GoodsType__Id", goodType.Id).OrderBy("Index")
		//获取所有文字商品
		qs.Filter("DisplayType", 0).All(&textGoods)
		rows["textGoods"] = textGoods
		//定义切片存取所有图片商品
		var imageGoods []models.IndexTypeGoodsBanner
		//获取所有图片商品
		qs.Filter("DisplayType", 1).All(&imageGoods)
		rows["imageGoods"] = imageGoods
		//将行容器存入总容器中
		goods = append(goods, rows)
	}
	//返回数据到前端
	this.Data["goods"] = goods
	//返回数据
	this.TplName = "index_sx.html"
}

/* 定义函数,负责商品详情页面展示 */
func (this *GoodsController) ShowGoodsDetail() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		fmt.Println("商品不存在!")
		this.Redirect("/index_sx", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goodsSKU models.GoodsSKU
	//goodsSKU.Id = id
	//o.Read(&goodsSKU)
	//上面查询属于单个查询,惰性查询,并不会查出其外键表中的数据,所以需要联合查询

	//获取商品详情
	qs := o.QueryTable("GoodsSKU")
	qs.RelatedSel("Goods", "GoodsType").Filter("Id", id).One(&goodsSKU)

	//获取同一类型的新品推荐
	var newGoods []models.GoodsSKU
	qs.RelatedSel("GoodsType").Filter("GoodsType__Name", goodsSKU.GoodsType.Name).OrderBy("-Time").Limit(2, 0).All(&newGoods)
	//返回数据
	fmt.Println()
	this.Data["newGoods"] = newGoods
	this.Data["goodsSKU"] = goodsSKU
	this.TplName = "detail.html"
}

/*
 定义函数,获取当前页码范围
 参1:总页码数;参2:当前页码数
*/
func GetPages(pageCount int, pageIndex int) []int {
	//定义切片,存入当前页码范围
	var pages []int
	//总页数小于5
	if pageCount <= 5 {
		for i := 1; i <= pageCount; i++ {
			pages = append(pages, i)
		}
	} else if pageIndex <= 3 {
		for i := 1; i <= 5; i++ {
			pages = append(pages, i)
		}
	} else if pageIndex >= pageCount-2 {
		for i := pageCount - 4; i <= pageCount; i++ {
			pages = append(pages, i)
		}
	} else {
		for i := pageIndex - 2; i <= pageIndex+2; i++ {
			pages = append(pages, i)
		}
	}
	//返回页码范围
	return pages
}

/* 定义函数,负责展示同一类所有商品 */
func (this *GoodsController) ShowTypeList() {
	//获取数据,类型id
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		fmt.Println("该类型不存在!")
		this.Redirect("/index_sx", 302)
		return
	}
	//处理数据
	//联合查询,查询该类型下的所有商品
	o := orm.NewOrm()
	var goodsSKUs []models.GoodsSKU
	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id)

	//列表页分页实现
	//获取总记录数
	count, _ := qs.Count()
	//定义每页记录数
	var pageSize int = 1
	//获取总页码
	pageCount := math.Ceil(float64(count) / float64(pageSize))
	//获取当前页数
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	//调用函数,获取当前页码范围
	pages := GetPages(int(pageCount), pageIndex)
	this.Data["pages"] = pages

	//实现上一页和下一页功能
	var prePage, nextPage int
	//判断范围
	if pageIndex-1 <= 0 {
		prePage = 1
	} else {
		prePage = pageIndex - 1
	}

	if pageIndex+1 >= int(pageCount) {
		nextPage = int(pageCount)
	} else {
		nextPage = pageIndex + 1
	}

	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage

	//查询当前页的数据记录
	qs = qs.Limit(pageSize, pageSize*(pageIndex-1))
	//列表页排序实现
	//获取排序字段
	sort := this.GetString("sort")
	if sort == "" { //sort=="",默认排序,正常查询
		qs.All(&goodsSKUs)
	} else if sort == "price" {
		qs.OrderBy("-Price").All(&goodsSKUs)
	} else {
		qs.OrderBy("-Sales").All(&goodsSKUs)
	}
	this.Data["sort"] = sort
	//返回数据
	this.Data["pageIndex"] = pageIndex
	this.Data["id"] = id
	this.Data["goodsSKUs"] = goodsSKUs
	this.TplName = "list.html"
}

/* 定义函数,负责页面搜索框搜索 */
func (this *GoodsController) HandleSearch() {
	//获取数据
	goodsName := this.GetString("goodsName")
	//校验数据
	if goodsName == ""{
		this.Redirect("/index",302)
	}
	//处理数据
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	//模糊查询,icontains-判断指定的字段是否包含第二个参数的值,i表示不区分大小写
	o.QueryTable("GoodsSKU").Filter("Name__icontains",goodsName).All(&goods)
	//返回数据
	this.Data["goods"] = goods
	this.TplName = "search.html"
}
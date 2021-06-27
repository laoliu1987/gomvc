package main

import (
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"gomvc/middlewares"
	"log"
	"net/http"
)


// 定义接收数据的结构体
type User struct {
	// binding:"required"修饰的字段，若接收为空值，则报错，是必须字段
	Uid int
	Name    string `form:"username" json:"username" uri:"user" xml:"user" binding:"required"`
	Password string `form:"password" json:"password" uri:"password" xml:"password" binding:"required"`

}

type Page struct {
	Page int `form:"page"`
	Size int `form:"size"`
	//Desc int `form:"desc"`
}

func main() {
	//读取db.ini配置文件
	cfg,err2 := goconfig.LoadConfigFile("./db.ini")
	if err2 != nil {
		log.Fatalf("无法加载配置文件：%s",err2)
	}

	db_username := cfg.MustValue("test","MYSQL_USER")
	db_password := cfg.MustValue("test","MYSQL_PASSWORD")
	db_address := cfg.MustValue("test","MYSQL_ADDRESS")
	db_database := cfg.MustValue("test","MYSQL_DATABASE")

	// 1.创建路由
	// 默认使用了2个中间件Logger(), Recovery()
	r := gin.Default()
	// 解决跨域问题
	r.Use(middlewares.Cors())

	// 连接mysql数据库
	//db,err1 := gorm.Open("mysql","root:322550@(127.0.0.1:3306)/info?charset=utf8&parseTime=True&loc=Local")
	db,err1 := gorm.Open("mysql",db_username+":"+db_password+"@("+db_address+")/"+db_database+"?charset=utf8&parseTime=True&loc=Local")
	if err1 != nil{
		panic(err1)
	}

	defer func(db *gorm.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	//模型与数据库表关联
	db.AutoMigrate(&User{})

	// 登录	&  JSON绑定
	r.POST("/login", func(c *gin.Context) {
		// 声明接收的变量
		var form User
		// Bind()默认解析并绑定form格式
		// 根据请求头中content-type自动推断

		if err := c.Bind(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 判断用户名密码是否正确
		if checkLogin(form,db) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"flag": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"flag": "ok"})
	})

	//分页查询

	r.GET("/member",func(c *gin.Context){

		var p Page

		users := make([]User,0)
		if c.ShouldBindQuery(&p) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "参数错误",
			})
			return
		}
		page := p.Page
		pageSize := p.Size
		//desc := p.Desc
		if page <= 0 {
			page = 1
		}

		var total int
		db.Model(&User{}).Count(&total)
		pageNum := total / pageSize
		if total % pageSize != 0{
			pageNum++
		}

		err := db.Limit(pageSize).Offset((page-1)*pageSize).Find(&users).Error
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  err.Error(),
			})
		}else{
			c.JSON(200, gin.H{
				"status":  "success",
				"data": users,
				"total": total,
				"page_num": pageNum,
			})
		}
	})

	//获取内存使用率

	r.GET("/getSysInfo",func(c *gin.Context){
		sysinfo := GetMemPercent()
		c.JSON(200,gin.H{
			"data":sysinfo,
		})
	})

	r.GET("/getDisk",func(c *gin.Context){
		sysinfo := GetDiskPercent()
		c.JSON(200,gin.H{
			"data":sysinfo,
		})
	})
	//启动服务器
	err := r.Run(":9003")
	if err != nil {
		return 
	}
}

func checkLogin(user User,db *gorm.DB) int {
	var count int
	db.Where("name = ? AND password = ?",user.Name,user.Password).Find(&user).Count(&count)
	fmt.Println(count)
	return count
}

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

func GetDiskPercent() int {
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	return int(diskInfo.UsedPercent)
}

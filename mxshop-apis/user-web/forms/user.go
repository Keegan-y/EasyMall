package forms

type PassWordLoginForm struct {
	Mobile    string `form:"mobile" json:"mobile" binding:"required,mobile"`
	PassWord  string `form:"password" json:"password" binding:"required,min=3,max=10"`
	Captcha   string `form:"captcha" json:"captcha" binding:"required,min=4,max=4"`
	CaptchaId string `form:"captcha_id" json:"captcha_id" binding:"required"`
}

type RegisterForm struct {
	Mobile    string `form:"mobile" json:"mobile" binding:"required,mobile"`
	PassWord  string `form:"password" json:"password" binding:"required,min=3,max=10"`
}
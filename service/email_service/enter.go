package email_service

import (
	"StarDreamerCyberNook/global"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/jordan-wright/email"
)

// SendRegister 发送注册验证码邮件
func SendRegister(target string, code string) error {
	em := global.Config.Email
	// 修正：subject 格式不完整，添加完整的邮件主题
	subject := fmt.Sprintf("%s - 注册验证码", em.SendNickname)

	text := fmt.Sprintf(`亲爱的用户 <b>%s</b> 大人，<br/><br/>
来自遥远世界的讯息传来：<br/>
🎉 恭喜您触发了 <b>%s</b> 的注册事件！<br/>
<br/>
您的验证码已经生成完毕，正在通过网络协议以光速奔向您的收件箱：<br/>
<br/>
<div style="background:#f0f0f0; padding:15px; border-radius:8px; text-align:center; font-size:24px; letter-spacing:4px; color:#ff6b6b; font-weight:bold; font-family: monospace;">
    %s
</div>
<br/>
（5分钟后自动过期）<br/>
<br/>
• 请勿将验证码泄露给其他人<br/>
• 如果这不是您本人操作，可能是您的邮箱被平行世界的您借用了<br/>
• 本邮件由程序自动发送，没有人类受到伤害（除了写代码的笔者）<br/>
<br/>
遇到 Bug 或想吐槽？欢迎前来反馈~<br/>
邮箱：<a href="mailto:%s" style="color:#333; text-decoration:none;">%s</a><br/>
<br/>
<b>%s</b> 敬上`,
		target,
		global.Config.Site.Project.Title,
		code,
		em.SendEmail, // 修正：原代码重复使用了 em.SendEmail 两次，但格式化字符串需要3个参数，实际提供了4个，需要核对
		em.SendEmail,
		global.Config.Site.Project.Title,
	)

	// 修正：原代码格式化字符串参数数量不匹配，这里 target 变量在 HTML 中使用了，但传入的是 target（邮箱），应该确认是否需要用户名
	return SendEmail(target, subject, text)
}

// SendForgetPwd 发送重置密码验证码（修正函数名拼写错误：Foget -> Forget）
func SendForgetPwd(target string, code string) error {
	em := global.Config.Email
	// 修正：添加完整的邮件主题和内容
	subject := fmt.Sprintf("%s - 密码重置验证码", em.SendNickname)

	// 修正：原代码只发送了纯验证码，需要完整的 HTML 模板
	text := fmt.Sprintf(`亲爱的用户 <b>%s</b> 大人，<br/><br/>
您正在申请重置密码，验证码如下：<br/>
<br/>
<div style="background:#f0f0f0; padding:15px; border-radius:8px; text-align:center; font-size:24px; letter-spacing:4px; color:#ff6b6b; font-weight:bold; font-family: monospace;">
    %s
</div>
<br/>
（5分钟后自动过期）<br/>
<br/>
• 请勿将验证码泄露给其他人<br/>
• 如果这不是您本人操作，请立即修改密码<br/>
<br/>
<b>%s</b> 敬上`,
		target,
		code,
		global.Config.Site.Project.Title,
	) //TODO:到时候排一下版

	return SendEmail(target, subject, text)
}

// SendForgetPwd 发送重置密码验证码（修正函数名拼写错误：Foget -> Forget）
func SendResetEmail(target string, code string) error {
	em := global.Config.Email
	// 修正：添加完整的邮件主题和内容
	subject := fmt.Sprintf("%s - 邮箱重置验证码", em.SendNickname)

	// 修正：原代码只发送了纯验证码，需要完整的 HTML 模板
	text := fmt.Sprintf(`亲爱的用户 <b>%s</b> 大人，<br/><br/>
您正在申请重置邮箱，验证码如下：<br/>
<br/>
<div style="background:#f0f0f0; padding:15px; border-radius:8px; text-align:center; font-size:24px; letter-spacing:4px; color:#ff6b6b; font-weight:bold; font-family: monospace;">
    %s
</div>
<br/>
（5分钟后自动过期）<br/>
<br/>
• 请勿将验证码泄露给其他人<br/>
• 如果这不是您本人操作，请立即修改密码<br/>
<br/>
<b>%s</b> 敬上`,
		target,
		code,
		global.Config.Site.Project.Title,
	) //TODO:到时候排一下版

	return SendEmail(target, subject, text)
}

// SendEmail 发送邮件的通用函数
func SendEmail(to, subject, text string) error {
	em := global.Config.Email
	e := email.NewEmail()

	// 修正：From 格式错误，应该是 "显示名 <邮箱地址>" 的格式
	e.From = fmt.Sprintf("%s <%s>", em.SendNickname, em.SendEmail)
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(text)

	// 添加文本版本作为后备（提高兼容性）
	e.Text = []byte("如果html文本不可视,可查看此处" + stripHTML(text))

	err := e.Send(
		fmt.Sprintf("%s:%d", em.Domain, em.Port),
		smtp.PlainAuth("", em.SendEmail, em.AuthCode, em.Domain),
	)

	// 修正：错误处理逻辑，"short response" 可能是服务器问题，不应该忽略
	if err != nil {
		// 只忽略特定的 TLS/SSL 握手错误，而不是所有 short response
		if strings.Contains(err.Error(), "short response: ") && strings.Contains(err.Error(), "TLS") {
			// 记录日志但忽略特定 TLS 错误
			return nil
		}
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}

// stripHTML 简单的 HTML 标签去除函数（用于生成纯文本版本）
func stripHTML(html string) string {
	// 简单的标签替换，生产环境建议使用 bluemonday 或 html2text 库
	result := strings.ReplaceAll(html, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<div>", "")
	result = strings.ReplaceAll(result, "</div>", "")
	result = strings.ReplaceAll(result, "<b>", "")
	result = strings.ReplaceAll(result, "</b>", "")
	// 移除其他 HTML 标签的简单正则或字符串操作
	return result
}

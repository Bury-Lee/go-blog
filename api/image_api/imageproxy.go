package image_api

//	type ImageproxyRequest struct {
//		Url string `json:"url" binding:"required"`
//	}
//
// ImageProxyView 图片代理转存接口
// 功能描述：
// 1. 接收客户端提供的图片远程 URL。
// 2. 后端发起请求下载该图片内容。
// 3. 校验图片大小是否超过配置限制。
// 4. 计算文件内容的 MD5 值作为唯一文件名（防止重复存储）。
// 5. 根据 Content-Type 确定文件后缀，将图片保存至本地服务器。
// 6. 返回保存后的本地访问路径。
//
// ⚠️ 风险与维护提示 (参考原代码 TODO)：
// - 法律风险：直接转存第三方图片可能涉及版权侵权问题，需确保业务场景合规（如用户明确授权或仅用于临时缓存）。
// - 必要性低：当前业务中转载文章场景较少，且现代 CDN 通常已处理防盗链，此接口的实际使用频率可能较低。
// - 资源消耗：大并发下会占用服务器带宽和磁盘 IO，建议评估是否保留或改为异步任务/纯代理模式（不落地存储）。
// - 数据一致性：目前图片虽已保存，但注释提到“图片还是要入库”，意味着数据库缺乏对应的记录，可能导致后期清理困难或无法在后台管理。
// func (ImageApi) ImageProxyView(c *gin.Context) {
// 	// 定义请求结构体，用于绑定参数（预期包含 Url 字段）
// 	var req ImageproxyRequest

// 	// 参数绑定与校验：自动解析 JSON/Form 参数到 req 结构体
// 	if err := c.ShouldBind(&req); err != nil {
// 		response.FailWithError(err, c)
// 		return
// 	}

// 	// 【关键步骤】发起 HTTP GET 请求下载远程图片
// 	// 注意：此处未设置超时时间，若远程服务器响应慢，可能导致当前协程长时间阻塞
// 	responseData, err := http.Get(req.Url)
// 	if err != nil {
// 		response.FailWithMsg("图片转存失败", c)
// 		return
// 	}
// 	// 确保在函数退出前关闭响应体，防止资源泄露
// 	defer responseData.Body.Close()

// 	// 读取全部响应数据到内存
// 	// ⚠️ 性能隐患：如果图片非常大，直接 ReadAll 会导致内存瞬间飙升，建议结合 io.LimitReader 或在流式处理中判断大小
// 	byteData, _ := io.ReadAll(responseData.Body)

// 	// 计算文件内容的 MD5 哈希值，用作文件名以实现去重
// 	hash := utils.Md5(byteData)

// 	// 获取配置中允许的最大上传文件大小 (单位: MB)
// 	s := global.Config.Upload.Size

// 	// 大小校验：如果文件字节数超过配置限制，拒绝保存
// 	if len(byteData) > int(s)*1024*1024 {
// 		response.FailWithMsg(fmt.Sprintf("文件大小大于%dMB,转存失败", s), c)
// 		return
// 	}

// 	// 默认文件后缀为 png
// 	suffix := "png"
// 	// 根据远程服务器返回的 Content-Type 动态调整后缀
// 	// 注意：目前只特殊处理了 avif，其他类型（如 jpeg, webp）仍会被强制存为 .png，可能导致文件头与扩展名不匹配
// 	switch responseData.Header.Get("Content-Type") {
// 	case "image/avif":
// 		suffix = "avif"
// 		// TODO: 建议补充对 image/jpeg, image/webp, image/gif 等常见格式的判断
// 	}

// 	// 构建本地存储路径：uploads/{上传目录}/{MD5哈希}.{后缀}
// 	filePath := fmt.Sprintf("uploads/%s/%s.%s", global.Config.Upload.UploadDir, hash, suffix)

// 	// 将二进制数据写入本地文件系统，权限设置为 0666 (实际权限受 umask 影响)
// 	err = os.WriteFile(filePath, byteData, 0666)
// 	if err != nil {
// 		logrus.Error(err)
// 		response.FailWithMsg("图片保存失败", c)
// 		return
// 	}

// 	// 返回成功响应，数据为图片的相对访问路径
// 	// 前端拿到此路径后，可将其作为本地资源引用，不再依赖原始 URL
// 	response.OkWithData("/"+filePath, c)
// }

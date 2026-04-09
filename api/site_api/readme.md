	/*
		如何获取请求体?c.Request.Body有阅后即焚的特性,无论是绑定还是字节流读取,在结束后都会销毁.
			解决方法:在读取后重新构造一个新的io.ReadCloser对象,并赋值给c.Request.Body
				byteData, err := io.ReadAll(c.Request.Body)
					if err != nil {
					  logrus.Errorf(err.Error())
					}
					fmt.Println("body: ", string(byteData))
					c.Request.Body = io.NopCloser(bytes.NewReader(byteData))
	*/
package context

type Context struct {
	*Config
	*Auth
}

var Ctx *Context

func init() {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}

	auth, err := NewAuth(config.ClientId, config.ClientSecret)
	if err != nil {
		panic(err)
	}

	Ctx = &Context{
		Config: config,
		Auth:   auth,
	}
}

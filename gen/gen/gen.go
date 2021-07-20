package gen

func Gen(args ...string) {
	switch args[0] {
	case "graphql":
		//graphql()
	case "rest":
		rest(args[1])
	}
}

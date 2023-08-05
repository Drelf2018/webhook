package data

type Comment struct {
	Post
	Repost string
}

type Comments struct {
	Root *Post
	Map  map[string]*Comment
}

func (cs *Comments) Query() {
	var r []Comment
	db.Find(&r, "platform = ?", cs.Root.Platform+cs.Root.Mid)
	for _, c := range r {
		cs.Map[c.Mid] = &c
	}
	for _, c := range cs.Map {
		cs.Insert(c)
	}
}

func (cs *Comments) Insert(c *Comment) {
	var root interface{ Insert(*Comment) } = cs.Root
	mc, ok := cs.Map[c.Repost]
	if ok {
		root = mc
	}
	root.Insert(c)
}

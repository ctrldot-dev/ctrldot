package bios

// NavItem implements list.DefaultItem for the left nav.
type NavItem struct {
	TitleVal string
	DescVal  string
}

func (n NavItem) Title() string       { return n.TitleVal }
func (n NavItem) Description() string { return n.DescVal }
func (n NavItem) FilterValue() string { return n.TitleVal }

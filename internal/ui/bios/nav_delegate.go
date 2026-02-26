package bios

import (
	"github.com/charmbracelet/bubbles/list"
)

// NewNavDelegate returns a list delegate styled with NavItem/NavSelected.
func NewNavDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	s := list.NewDefaultItemStyles()
	s.NormalTitle = NavItemStyle
	s.SelectedTitle = NavSelectedStyle
	s.DimmedTitle = Muted
	s.NormalDesc = Muted
	s.SelectedDesc = Muted
	s.DimmedDesc = Muted
	s.FilterMatch = NavItemStyle
	d.Styles = s
	d.SetSpacing(0)
	return d
}

// StyledListStyles returns list.Styles so no default colours leak; use with list.Styles = StyledListStyles().
func StyledListStyles() list.Styles {
	s := list.DefaultStyles()
	s.Title = Title
	s.PaginationStyle = Muted
	s.HelpStyle = Muted
	s.TitleBar = StatusBar
	return s
}

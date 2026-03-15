package model

// BrandingSetting stores runtime brand copy, theme, and asset settings managed from admin.
type BrandingSetting struct {
	BaseModel
	AppTitle       string `gorm:"size:128" json:"appTitle"`
	ConsoleName    string `gorm:"size:128" json:"consoleName"`
	ProductTagline string `gorm:"size:255" json:"productTagline"`
	LogoMarkURL    string `gorm:"size:255" json:"logoMarkUrl"`
	FaviconURL     string `gorm:"size:255" json:"faviconUrl"`
	LoginHeroURL   string `gorm:"size:255" json:"loginHeroUrl"`
	Primary        string `gorm:"size:16" json:"primary"`
	PrimaryDark    string `gorm:"size:16" json:"primaryDark"`
	ShellStart     string `gorm:"size:16" json:"shellStart"`
	ShellEnd       string `gorm:"size:16" json:"shellEnd"`
	HeroStart      string `gorm:"size:16" json:"heroStart"`
	HeroEnd        string `gorm:"size:16" json:"heroEnd"`
}

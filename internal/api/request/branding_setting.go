package request

// BrandingThemeRequest describes theme colors submitted from admin.
type BrandingThemeRequest struct {
	Primary     string `json:"primary" binding:"omitempty,max=16"`
	PrimaryDark string `json:"primaryDark" binding:"omitempty,max=16"`
	ShellStart  string `json:"shellStart" binding:"omitempty,max=16"`
	ShellEnd    string `json:"shellEnd" binding:"omitempty,max=16"`
	HeroStart   string `json:"heroStart" binding:"omitempty,max=16"`
	HeroEnd     string `json:"heroEnd" binding:"omitempty,max=16"`
}

// UpdateBrandingSettingRequest describes branding updates from admin.
type UpdateBrandingSettingRequest struct {
	AppTitle       string               `json:"appTitle" binding:"omitempty,max=128"`
	ConsoleName    string               `json:"consoleName" binding:"omitempty,max=128"`
	ProductTagline string               `json:"productTagline" binding:"omitempty,max=255"`
	LogoMarkURL    string               `json:"logoMarkUrl" binding:"omitempty,max=255"`
	LoginHeroURL   string               `json:"loginHeroUrl" binding:"omitempty,max=255"`
	Theme          BrandingThemeRequest `json:"theme"`
}

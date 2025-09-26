package structs

type (
	User struct {
		ID         uint64     `json:"id" gorm:"primaryKey"`
		Email      string     `json:"email" gorm:"unique"` //
		VarifEmail bool       `json:"varifemail" `         //
		Password   string     `gorm:"->:false;<-:create"`
		Created    uint64     `json:"created" gorm:"autoCreateTime"`
		Updated    uint64     `json:"updated" gorm:"autoUpdateTime"`
		ParamsUser ParamsUser `json:"params" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
		AuthUser   AuthUser   `json:"auth" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}
	UserShow struct { // структура используется только для вывода данных пользователя
		ID    uint64 // `json:"id" gorm:"primaryKey"`
		Email string // `json:"email" gorm:"unique"`
		Name  string // `json:"name" gorm:"default:''"`       // name of user
		Root  uint8  // `json:"root" gorm:"default=0;size=1"` // доступ (0-guest 1-simpleuser 3-premiumuser 5-moderator 7-admin 9-superadmin)
	}
	ParamsUser struct {
		ID     uint64 `json:"id" gorm:"primaryKey"`
		UserID uint64 `json:"uid" gorm:"unique, foreignKey"` // userID
		Name   string `json:"name" gorm:"default:''"`        // name of user
		Root   uint8  `json:"root" gorm:"default=0; size=1"` // доступ (0-guest 1-simpleuser 3-premiumuser 5-moderator 7-admin 9-superadmin)
	}

	AuthUser struct {
		ID              uint64 `json:"id" gorm:"primaryKey"`
		UserID          uint64 `json:"uid" gorm:"unique, foreignKey"` // userID
		ExpareTime      int64  `json:"updated" gorm:"autoUpdateTime"` // unix expare time for refreshToken
		LastRefreshTime int64  `json:"lastrefreshtime"`               // unix последней выдачи пользователю рефреш токена (по факту дата последнего визита)
		LastEntryRoot   uint8  `json:"lastentryroot"`                 // root пользователя во время последней авторизации (через стандартный вход)
		RefreshToken    string `json:"refreshToken"`                  // jwt
	}
	UserMailConfirm struct {
		ID               uint64 `json:"id" gorm:"primaryKey"`
		UserID           uint64 `json:"uid" gorm:"unique, foreignKey"` // userID
		ExpareTime       int64  `json:"updated" gorm:"autoUpdateTime"` // expare time for varification code
		VarificationCode string `json:"varifCode"`                     // code like string length 25 simbols
	}
)

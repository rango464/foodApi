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
		Root  uint8  // `json:"root" gorm:"default=1;size=1"` // доступ (0-superadmin 1-admin 3-moderator 5-superuser 7-simpleuser 9-guest)
	}
	ParamsUser struct {
		ID     uint64 `json:"id" gorm:"primaryKey"`
		UserID uint64 `json:"uid" gorm:"unique, foreignKey"` // userID
		Name   string `json:"name" gorm:"default:''"`        // name of user
		Root   uint8  `json:"root" gorm:"default=9;size=1"`  // доступ (0-superadmin 1-admin 3-moderator 5-superuser 7-simpleuser 9-guest)
	}
	AuthUser struct {
		ID           uint64 `json:"id" gorm:"primaryKey"`
		UserID       uint64 `json:"uid" gorm:"unique, foreignKey"` // userID
		ExpareTime   int64  `json:"updated" gorm:"autoUpdateTime"` // expare time for refreshToken
		RefreshToken string `json:"refreshToken"`                  // jwt
	}
	UserMailConfirm struct {
		ID               uint64 `json:"id" gorm:"primaryKey"`
		UserID           uint64 `json:"uid" gorm:"unique, foreignKey"` // userID
		ExpareTime       int64  `json:"updated" gorm:"autoUpdateTime"` // expare time for varification code
		VarificationCode string `json:"varifCode"`                     // code like string length 25 simbols
	}
)

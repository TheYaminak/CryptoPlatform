package models

type User struct {
	Name     string `json:Name`
	LastName string `json:LastName`
	Email    string `json:Email`
	Password string `json:Password`
}

type Login struct {
	Email    string `json:Email`
	Password string `json:Password`
}

type Billto struct {
	FirstName      string `json:FirstName`
	LastName       string `json:LastName`
	Address        string `json:Address`
	City           string `json:City`
	Province       string `json:Province`
	Country        string `json:Country`
	PhoneNumber    string `json:PhoneNumber`
	CardNumber     string `json:CardNumber`
	ExpirationDate string `json:ExpirationDate`
	CardCode       string `json:CardCode`
	RNC            string `json:RNC`
	Amount         string `json:Amount`
	Email          string `json:Email`
	TypeRNC        string `json:TypeRNC`
	//Debes de poner el tipo de RNC 1 consumidor final 2
}

type City struct {
	Id       string `json:Id`
	Name     string `json:Name`
	Province string `json:Province`
}

type Province struct {
	Id   string `json:Id`
	Name string `json:Name`
}

type CitiesAndProvinces struct {
	Province struct {
		Id   string `json:Id`
		Name string `json:Name`
	}
	City struct {
		Id       string `json:Id`
		Name     string `json:Name`
		Province string `json:Province`
	}
}

type PaymentSettings struct {
	Id     string `json:Id`
	Number string `json:Number`
	Cvv    string `json:Cvv`
	Date   string `json:Date`
}

type ProfileSetting struct {
	Email    string `json:Email`
	Name     string `json:Name`
	LastName string `json:Lastname`
	RNC      string `json:RNC`
	Address  string `json:Address`
	City     string `json:City`
	Phone    string `json:Phone`
}

type UserSettings struct {
	Email    string `json:Email`
	Name     string `json:Name`
	LastName string `json:Lastname`
	RNC      string `json:RNC`
	TypeRNC  string `json:TypeRNC`
}

type ContactSettings struct {
	Address     string `json:Address`
	City        string `json:City`
	PhoneNumber string `json:PhoneNumber`
}

type RNC struct {
	Min        string `json:Min`
	Max        string `json:Max`
	Type       string `json:Type`
	Expiration string `json:Expiration`
}

type SelectRnc struct {
	Rnc   string `json:Rnc`
	TypeR string `json:TypeR`
}

type Vouchers struct {
	Name        string `json:Name`
	Date        string `json:Date`
	Description string `json:Description`
	Subtotal    string `json:Subtotal`
	RNC         string `json:RNC`
	Voucher     string `json:Voucher`
}

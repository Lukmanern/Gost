package model

type UserCreate struct {
	Name     string `validate:"required,string,min=5,max=60" json:"name"`
	Email    string `validate:"required,string,min=5,max=60" json:"email"`
	Password string `validate:"required,string,min=8,max=30" json:"password"`
}

type UserResponse struct {
	ID    int    `validate:"required,int,min=1"`
	Name  string `validate:"required,string,min=5,max=60" json:"name"`
	Email string `validate:"required,string,min=5,max=60" json:"email"`
}

type UserLogin struct {
	Email    string `validate:"required,string,min=5,max=60" json:"email"`
	Password string `validate:"required,string,min=8,max=30" json:"password"`
}

type UserUpdate struct {
	ID   int    `validate:"required,int,min=1"`
	Name string `validate:"required,string,min=5,max=60" json:"name"`
	// Password string `validate:"required,string,min=8,max=30" json:"password"`
}

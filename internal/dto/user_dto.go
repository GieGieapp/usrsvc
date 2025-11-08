package dto

type CreateCustomerRequest struct {
	CstName       string `json:"cst_name" validate:"required"`
	CstDob        string `json:"cst_dob" validate:"required"`
	NationalityID int32  `json:"nationality_id" validate:"required,gt=0"`
	CstPhoneNum   string `json:"cst_phoneNum" validate:"required"`
	CstEmail      string `json:"cst_email" validate:"required,email"`
	Family        []struct {
		FlRelation string `json:"fl_relation" validate:"required"`
		FlName     string `json:"fl_name"    validate:"required"`
		FlDob      string `json:"fl_dob"     validate:"required"`
	} `json:"family"`
}
type UpdateCustomerRequest = CreateCustomerRequest

package dto

type CustomerListItem struct {
	CstID         int32  `json:"cst_id"`
	CstName       string `json:"cst_name"`
	CstDob        string `json:"cst_dob"`
	NationalityID int32  `json:"nationality_id"`
	CstPhoneNum   string `json:"cst_phoneNum"`
	CstEmail      string `json:"cst_email"`
}

type CustomerListResponse struct {
	Data  []CustomerListItem `json:"data"`
	Total int                `json:"total"`
}

package dto

type FamilyMemberResponse struct {
	FlRelation string `json:"fl_relation"`
	FlName     string `json:"fl_name"`
	FlDob      string `json:"fl_dob"`
}

type CustomerResponse struct {
	CstID         int32                  `json:"cst_id"`
	CstName       string                 `json:"cst_name"`
	CstDob        string                 `json:"cst_dob"`
	NationalityID int32                  `json:"nationality_id"`
	CstPhoneNum   string                 `json:"cst_phoneNum"`
	CstEmail      string                 `json:"cst_email"`
	Family        []FamilyMemberResponse `json:"family"`
}

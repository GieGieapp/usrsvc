package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"usrsvc/internal/pkg/log"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"usrsvc/internal/domain"
	"usrsvc/internal/dto"
)

type Handler struct {
	UC  domain.UserUsecase
	Val *validator.Validate
}

func NewHandler(uc domain.UserUsecase) *Handler { return &Handler{UC: uc, Val: validator.New()} }

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	search := q.Get("search")

	rows, total, err := h.UC.List(r.Context(), search, page, size)
	if err != nil {
		log.Error.Printf("list_users repo_err err=%v", err)
		writeErr(w, StatusInternalServerError, MsgInternal, nil)
		return
	}

	out := make([]dto.CustomerListItem, 0, len(rows))
	for _, c := range rows {
		out = append(out, dto.CustomerListItem{
			CstID:         c.ID,
			CstName:       strings.TrimSpace(c.Name),
			CstDob:        c.Dob.Format("2006-01-02"),
			NationalityID: c.NationalityID,
			CstPhoneNum:   c.PhoneNum,
			CstEmail:      c.Email,
		})
	}
	log.Info.Printf("list_users ok total=%d", total)
	writeJSON(w, StatusOK, dto.CustomerListResponse{Data: out, Total: int(total)})
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if id == 0 {
		log.Error.Printf("get_user invalid_id")
		writeErr(w, StatusBadRequest, MsgInvalidID, nil)
		return
	}
	c, err := h.UC.Get(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeErr(w, StatusNotFound, MsgNotFound, nil)
			return
		}
		log.Error.Printf("get_user repo_err id=%d err=%v", id, err)
		writeErr(w, StatusInternalServerError, MsgInternal, nil)
		return
	}
	if c == nil {
		writeErr(w, StatusNotFound, MsgNotFound, nil)
		return
	}
	log.Info.Printf("get_user ok id=%d", id)
	writeJSON(w, StatusOK, c)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error.Printf("create_user decode_json err=%v", err)
		writeErr(w, StatusBadRequest, MsgInvalidJSON, nil)
		return
	}
	if err := h.Val.Struct(req); err != nil {
		log.Error.Printf("create_user validate err=%v body=%+v", err, req)
		writeErr(w, StatusUnprocessableEntity, MsgValidation, nil)
		return
	}
	if _, err := time.Parse("2006-01-02", req.CstDob); err != nil {
		log.Error.Printf("create_user bad_dob err=%v dob=%s", err, req.CstDob)
		writeErr(w, StatusUnprocessableEntity, "invalid cst_dob", map[string]string{"cst_dob": "YYYY-MM-DD"})
		return
	}
	for i, f := range req.Family {
		if _, err := time.Parse("2006-01-02", f.FlDob); err != nil {
			log.Error.Printf("create_user bad_family_dob idx=%d err=%v dob=%s", i, err, f.FlDob)
			writeErr(w, StatusUnprocessableEntity, "invalid fl_dob", map[string]string{"family[" + strconv.Itoa(i) + "].fl_dob": "YYYY-MM-DD"})
			return
		}
	}

	c := domain.Customer{
		NationalityID: req.NationalityID,
		Name:          req.CstName,
		Dob:           mustParse(req.CstDob),
		PhoneNum:      req.CstPhoneNum,
		Email:         req.CstEmail,
	}
	for _, f := range req.Family {
		c.Family = append(c.Family, domain.FamilyMember{
			Relation: f.FlRelation, Name: f.FlName, Dob: mustParse(f.FlDob),
		})
	}

	id, err := h.UC.Create(r.Context(), c)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			log.Info.Printf("create_user conflict email=%q", c.Email)
			writeErr(w, StatusConflict, MsgConflict, map[string]string{"cst_email": "already exists"})
			return
		}
		log.Error.Printf("create_user repo_err name=%q email=%q err=%v", c.Name, c.Email, err)
		writeErr(w, StatusInternalServerError, MsgInternal, nil)
		return
	}

	resp := dto.CustomerResponse{
		CstID:         id,
		CstName:       c.Name,
		CstDob:        c.Dob.Format("2006-01-02"),
		NationalityID: c.NationalityID,
		CstPhoneNum:   c.PhoneNum,
		CstEmail:      c.Email,
		Family:        make([]dto.FamilyMemberResponse, 0, len(c.Family)),
	}
	for _, f := range c.Family {
		resp.Family = append(resp.Family, dto.FamilyMemberResponse{
			FlRelation: f.Relation,
			FlName:     f.Name,
			FlDob:      f.Dob.Format("2006-01-02"),
		})
	}

	log.Info.Printf("create_user ok id=%d name=%q email=%q family=%d", id, c.Name, c.Email, len(c.Family))
	writeJSON(w, StatusCreated, resp)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if id == 0 {
		log.Error.Printf("update_user invalid_id")
		writeErr(w, StatusBadRequest, MsgInvalidID, nil)
		return
	}
	var req dto.UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error.Printf("update_user decode_json err=%v", err)
		writeErr(w, StatusBadRequest, MsgInvalidJSON, nil)
		return
	}
	if err := h.Val.Struct(req); err != nil {
		log.Error.Printf("update_user validate err=%v body=%+v", err, req)
		writeErr(w, StatusUnprocessableEntity, MsgValidation, nil)
		return
	}
	if _, err := time.Parse("2006-01-02", req.CstDob); err != nil {
		log.Error.Printf("update_user bad_dob err=%v dob=%s", err, req.CstDob)
		writeErr(w, StatusUnprocessableEntity, "invalid cst_dob", map[string]string{"cst_dob": "YYYY-MM-DD"})
		return
	}

	c := domain.Customer{
		NationalityID: req.NationalityID,
		Name:          req.CstName,
		Dob:           mustParse(req.CstDob),
		PhoneNum:      req.CstPhoneNum,
		Email:         req.CstEmail,
	}
	for i, f := range req.Family {
		if _, err := time.Parse("2006-01-02", f.FlDob); err != nil {
			log.Error.Printf("update_user bad_family_dob idx=%d err=%v dob=%s", i, err, f.FlDob)
			writeErr(w, StatusUnprocessableEntity, "invalid fl_dob", map[string]string{"family[" + strconv.Itoa(i) + "].fl_dob": "YYYY-MM-DD"})
			return
		}
		c.Family = append(c.Family, domain.FamilyMember{
			Relation: f.FlRelation, Name: f.FlName, Dob: mustParse(f.FlDob),
		})
	}

	if err := h.UC.Update(r.Context(), int32(id), c); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeErr(w, StatusNotFound, MsgNotFound, nil)
			return
		}
		log.Error.Printf("update_user repo_err id=%d name=%q email=%q err=%v", id, c.Name, c.Email, err)
		writeErr(w, StatusInternalServerError, MsgInternal, nil)
		return
	}
	log.Info.Printf("update_user ok id=%d family=%d", id, len(c.Family))
	writeJSON(w, StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		log.Error.Printf("delete_user invalid_id id=%q", idStr)
		writeErr(w, StatusBadRequest, MsgInvalidID, nil)
		return
	}
	if err := h.UC.Delete(r.Context(), int32(id)); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeErr(w, StatusNotFound, MsgNotFound, nil)
			return
		}
		log.Error.Printf("delete_user repo_err id=%d err=%v", id, err)
		writeErr(w, StatusInternalServerError, MsgInternal, nil)
		return
	}
	log.Info.Printf("delete_user ok id=%d", id)
	writeJSON(w, StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListNationality(w http.ResponseWriter, r *http.Request) {
	n, err := h.UC.ListNationality(r.Context())
	if err != nil {
		log.Error.Printf("list_nationality repo_err err=%v", err)
		writeErr(w, StatusInternalServerError, MsgInternal, nil)
		return
	}
	if n == nil {
		n = []domain.Nationality{}
	}
	log.Info.Printf("list_nationality ok count=%d", len(n))
	writeJSON(w, StatusOK, n)
}

func mustParse(s string) (t time.Time) { t, _ = time.Parse("2006-01-02", s); return }

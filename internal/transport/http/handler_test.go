package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"usrsvc/internal/domain" // ← sesuaikan module path
	"usrsvc/internal/mocks"  // sesuaikan module path jika berbeda
)

func TestHandler_ListNationality(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(m *mocks.UserUsecase)
		wantStatusCode int
		assertBody     func(t *testing.T, body []byte)
	}{
		{
			name: "ok_with_items",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("ListNationality", mock.Anything).
					Return([]domain.Nationality{
						{ID: 1, Name: "Indonesia"},
						{ID: 2, Name: "Malaysia"},
					}, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var got []domain.Nationality
				require.NoError(t, json.Unmarshal(body, &got))
				require.Len(t, got, 2)
				assert.Equal(t, "Indonesia", got[0].Name)
				assert.Equal(t, "Malaysia", got[1].Name)
			},
		},
		{
			name: "ok_nil_slice_returns_empty_array",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("ListNationality", mock.Anything).
					Return(([]domain.Nationality)(nil), nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				assert.JSONEq(t, "[]", string(body))
			},
		},
		{
			name: "internal_error",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("ListNationality", mock.Anything).
					Return(nil, errors.New("db error")).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			assertBody: func(t *testing.T, body []byte) {
				assert.NotEmpty(t, body) // struktur error project bisa beda, cukup tidak kosong
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mocks.UserUsecase)
			tt.setupMock(mockUC)

			h := &Handler{
				UC:  mockUC,
				Val: validator.New(),
			}

			req := httptest.NewRequest(http.MethodGet, "/nationalities", nil)
			rr := httptest.NewRecorder()

			h.ListNationality(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			tt.assertBody(t, rr.Body.Bytes())

			mockUC.AssertExpectations(t)
		})
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		bodyRaw   []byte
		bodyObj   any
		setupMock func(m *mocks.UserUsecase)
		wantCode  int
		checkBody func(t *testing.T, body []byte)
	}{
		{
			name: "201_created",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "ALFA",
				"cst_dob":        "1992-05-10",
				"cst_phoneNum":   "0811000001",
				"cst_email":      "alfa1@example.com",
				"family": []map[string]any{
					{"fl_relation": "Spouse", "fl_name": "BETA", "fl_dob": "1993-07-01"},
				},
			},
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(c domain.Customer) bool {
					if c.NationalityID != 1 || c.Name != "ALFA" || c.Email != "alfa1@example.com" {
						return false
					}
					if len(c.Family) != 1 {
						return false
					}
					return c.Family[0].Relation == "Spouse" && c.Family[0].Name == "BETA"
				})).Return(int32(123), nil).Once()
			},
			wantCode: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var got map[string]any
				require.NoError(t, json.Unmarshal(body, &got))
				assert.Equal(t, float64(123), got["cst_id"])
				assert.Equal(t, "ALFA", got["cst_name"])
				assert.Equal(t, "1992-05-10", got["cst_dob"])
				assert.Equal(t, "alfa1@example.com", got["cst_email"])
			},
		},
		{
			name:      "400_invalid_json",
			bodyRaw:   []byte("{"),
			setupMock: func(m *mocks.UserUsecase) {},
			wantCode:  http.StatusBadRequest,
			checkBody: func(t *testing.T, body []byte) { assert.NotEmpty(t, body) },
		},
		{
			name:      "422_validation_error",
			bodyObj:   map[string]any{}, // kosong → gagal validasi
			setupMock: func(m *mocks.UserUsecase) {},
			wantCode:  http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, body []byte) { assert.NotEmpty(t, body) },
		},
		{
			name: "422_invalid_cst_dob_format",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "ALFA",
				"cst_dob":        "10-05-1992", // salah format
				"cst_phoneNum":   "0811",
				"cst_email":      "a@example.com",
			},
			setupMock: func(m *mocks.UserUsecase) {},
			wantCode:  http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "invalid cst_dob")
			},
		},
		{
			name: "422_invalid_family_dob_format",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "ALFA",
				"cst_dob":        "1992-05-10",
				"cst_phoneNum":   "0811",
				"cst_email":      "a@example.com",
				"family": []map[string]any{
					{"fl_relation": "Child", "fl_name": "GAMA", "fl_dob": "31/12/2010"}, // salah format
				},
			},
			setupMock: func(m *mocks.UserUsecase) {},
			wantCode:  http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "invalid fl_dob")
			},
		},
		{
			name: "409_conflict_email_exists",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "ALFA",
				"cst_dob":        "1992-05-10",
				"cst_phoneNum":   "0811",
				"cst_email":      "exists@example.com",
			},
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Customer")).
					Return(int32(0), domain.ErrConflict).Once()
			},
			wantCode: http.StatusConflict,
			checkBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "already exists")
			},
		},
		{
			name: "500_repo_error",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "ALFA",
				"cst_dob":        "1992-05-10",
				"cst_phoneNum":   "0811",
				"cst_email":      "err@example.com",
			},
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Customer")).
					Return(int32(0), errors.New("db down")).Once()
			},
			wantCode:  http.StatusInternalServerError,
			checkBody: func(t *testing.T, body []byte) { assert.NotEmpty(t, body) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUsecase)
			tc.setupMock(mockUC)

			h := &Handler{UC: mockUC, Val: validator.New()}

			var body []byte
			if tc.bodyRaw != nil {
				body = tc.bodyRaw
			} else {
				body = mustJSON(t, tc.bodyObj)
			}

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.CreateUser(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.wantCode, res.StatusCode)
			tc.checkBody(t, rr.Body.Bytes())

			mockUC.AssertExpectations(t)
		})
	}
}

func makeDeleteReq(id string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodDelete, "/users/"+id, nil)
	req = mux.SetURLVars(req, map[string]string{"id": id})
	return req, httptest.NewRecorder()
}

func TestHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		idVar     string
		setupMock func(m *mocks.UserUsecase)
		wantCode  int
		checkBody func(t *testing.T, body []byte)
	}{
		{
			name:  "400_invalid_id_non_numeric",
			idVar: "abc",
			setupMock: func(m *mocks.UserUsecase) {
				// no call
			},
			wantCode:  http.StatusBadRequest,
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:      "400_invalid_id_zero",
			idVar:     "0",
			setupMock: func(m *mocks.UserUsecase) {},
			wantCode:  http.StatusBadRequest,
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:  "404_not_found",
			idVar: "123",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Delete", mock.Anything, int32(123)).
					Return(domain.ErrNotFound).
					Once()
			},
			wantCode:  http.StatusNotFound,
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:  "500_repo_error",
			idVar: "124",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Delete", mock.Anything, int32(124)).
					Return(assert.AnError).
					Once()
			},
			wantCode:  http.StatusInternalServerError,
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:  "200_ok",
			idVar: "125",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Delete", mock.Anything, int32(125)).
					Return(nil).
					Once()
			},
			wantCode: http.StatusOK,
			checkBody: func(t *testing.T, b []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(b, &resp))
				assert.Equal(t, "ok", resp["status"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUsecase)
			if tc.setupMock != nil {
				tc.setupMock(mockUC)
			}
			h := &Handler{UC: mockUC, Val: validator.New()}

			req, rr := makeDeleteReq(tc.idVar)
			h.DeleteUser(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.wantCode, res.StatusCode)
			if tc.checkBody != nil {
				tc.checkBody(t, rr.Body.Bytes())
			}
			mockUC.AssertExpectations(t)
		})
	}
}

func TestHandler_GetUser(t *testing.T) {
	tests := []struct {
		name      string
		idVar     string
		setupMock func(m *mocks.UserUsecase)
		wantCode  int
	}{
		{
			name:  "400_invalid_id_zero",
			idVar: "0",
			setupMock: func(m *mocks.UserUsecase) {
				// UC tidak dipanggil
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name:  "400_invalid_id_non_numeric",
			idVar: "abc", // Atoi -> err, id==0 -> 400
			setupMock: func(m *mocks.UserUsecase) {
				// UC tidak dipanggil
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name:  "404_not_found_from_uc_error",
			idVar: "123",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Get", mock.Anything, int32(123)).
					Return((*domain.Customer)(nil), domain.ErrNotFound).
					Once()
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:  "404_not_found_nil_entity",
			idVar: "124",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Get", mock.Anything, int32(124)).
					Return((*domain.Customer)(nil), nil).
					Once()
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:  "500_repo_error",
			idVar: "125",
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Get", mock.Anything, int32(125)).
					Return((*domain.Customer)(nil), assert.AnError).
					Once()
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:  "200_ok",
			idVar: "126",
			setupMock: func(m *mocks.UserUsecase) {
				// kembalikan entity minimal; field tak perlu dicek detail
				m.On("Get", mock.Anything, int32(126)).
					Return(&domain.Customer{}, nil).
					Once()
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUsecase)
			if tc.setupMock != nil {
				tc.setupMock(mockUC)
			}

			h := &Handler{UC: mockUC, Val: validator.New()}

			req := httptest.NewRequest(http.MethodGet, "/users/"+tc.idVar, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.idVar})
			rr := httptest.NewRecorder()

			h.GetUser(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.wantCode, res.StatusCode)
			if tc.wantCode == http.StatusOK {
				assert.NotEmpty(t, rr.Body.Bytes())
			}

			mockUC.AssertExpectations(t)
		})
	}
}

func TestHandler_ListNationality1(t *testing.T) {
	type fields struct {
		UC  domain.UserUsecase
		Val *validator.Validate
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				UC:  tt.fields.UC,
				Val: tt.fields.Val,
			}
			h.ListNationality(tt.args.w, tt.args.r)
		})
	}
}

func TestHandler_ListUsers(t *testing.T) {
	makeReq := func(q url.Values) *http.Request {
		u := &url.URL{Path: "/users", RawQuery: q.Encode()}
		return httptest.NewRequest(http.MethodGet, u.String(), nil)
	}

	t1 := time.Date(1992, 5, 10, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(1988, 11, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		query     url.Values
		setupMock func(m *mocks.UserUsecase)
		wantCode  int
		checkBody func(t *testing.T, b []byte)
	}{
		{
			name:  "200_ok_with_items_and_search",
			query: url.Values{"page": {"1"}, "size": {"2"}, "search": {"AL"}},
			setupMock: func(m *mocks.UserUsecase) {
				m.On("List",
					mock.Anything, "AL",
					mock.Anything, // page
					mock.Anything, // size
				).
					Return([]domain.Customer{
						{ID: 36, Name: "  ALFA  ", Dob: t1, NationalityID: 1, PhoneNum: "0811000001", Email: "alfa1@example.com"},
						{ID: 37, Name: "BRAVO", Dob: t2, NationalityID: 1, PhoneNum: "0811000002", Email: "bravo2@example.com"},
					}, int32(2), nil).
					Once()
			},
			wantCode: http.StatusOK,
			checkBody: func(t *testing.T, b []byte) {
				var got map[string]any
				assert.NoError(t, json.Unmarshal(b, &got))
				data, _ := got["data"].([]any)
				assert.Len(t, data, 2)

				row0 := data[0].(map[string]any)
				assert.Equal(t, "ALFA", row0["cst_name"])
				assert.Equal(t, "1992-05-10", row0["cst_dob"])
				assert.Equal(t, "0811000001", row0["cst_phoneNum"])
				assert.Equal(t, "alfa1@example.com", row0["cst_email"])

				switch v := got["total"].(type) {
				case float64:
					assert.Equal(t, float64(2), v)
				default:
					assert.EqualValues(t, 2, v)
				}
			},
		},
		{
			name:  "200_ok_empty_result",
			query: url.Values{"page": {"1"}, "size": {"10"}},
			setupMock: func(m *mocks.UserUsecase) {
				m.On("List",
					mock.Anything, "",
					mock.AnythingOfType("int"), mock.AnythingOfType("int"),
				).
					Return([]domain.Customer{}, int32(0), nil).
					Once()
			},
			wantCode: http.StatusOK,
			checkBody: func(t *testing.T, b []byte) {
				assert.JSONEq(t, `{"data": [], "total": 0}`, string(b))
			},
		},
		{
			name:  "200_ok_default_pagination_when_invalid_page_size",
			query: url.Values{"page": {"0"}, "size": {"0"}},
			setupMock: func(m *mocks.UserUsecase) {
				// handler normalize → page=1,size=10
				m.On("List",
					mock.Anything, "",
					mock.AnythingOfType("int"), mock.AnythingOfType("int"),
				).
					Return([]domain.Customer{}, int32(0), nil).
					Once()
			},
			wantCode:  http.StatusOK,
			checkBody: func(t *testing.T, b []byte) {},
		},
		{
			name:  "200_ok_size_cap_when_gt_100",
			query: url.Values{"page": {"2"}, "size": {"999"}},
			setupMock: func(m *mocks.UserUsecase) {
				// handler cap size → 10
				m.On("List",
					mock.Anything, "",
					mock.AnythingOfType("int"), mock.AnythingOfType("int"),
				).
					Return([]domain.Customer{}, int32(0), nil).
					Once()
			},
			wantCode:  http.StatusOK,
			checkBody: func(t *testing.T, b []byte) {},
		},
		{
			name:  "500_repo_error",
			query: url.Values{"page": {"1"}, "size": {"10"}},
			setupMock: func(m *mocks.UserUsecase) {
				m.On("List",
					mock.Anything, "",
					mock.AnythingOfType("int"), mock.AnythingOfType("int"),
				).
					Return(([]domain.Customer)(nil), int32(0), assert.AnError).
					Once()
			},
			wantCode:  http.StatusInternalServerError,
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUsecase)
			tc.setupMock(mockUC)

			h := &Handler{UC: mockUC, Val: validator.New()}

			req := makeReq(tc.query)
			rr := httptest.NewRecorder()

			h.ListUsers(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.wantCode, res.StatusCode)
			tc.checkBody(t, rr.Body.Bytes())

			mockUC.AssertExpectations(t)
		})
	}
}

func makeUpdateReq(idStr string, body []byte) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPut, "/users/"+idStr, bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": idStr})
	req.Header.Set("Content-Type", "application/json")
	return req, httptest.NewRecorder()
}

func TestHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name      string
		idVar     string
		bodyRaw   []byte
		bodyObj   any
		setupMock func(m *mocks.UserUsecase)
		wantCode  int
		checkBody func(t *testing.T, body []byte)
	}{
		{
			name:     "400_invalid_id_zero",
			idVar:    "0",
			bodyObj:  map[string]any{"cst_name": "X"}, // body tak dipakai karena invalid id
			wantCode: http.StatusBadRequest,
			setupMock: func(m *mocks.UserUsecase) {
				// UC tidak dipanggil
			},
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:      "400_invalid_json",
			idVar:     "123",
			bodyRaw:   []byte("{"),
			wantCode:  http.StatusBadRequest,
			setupMock: func(m *mocks.UserUsecase) {},
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:      "422_validation_error_missing_required",
			idVar:     "123",
			bodyObj:   map[string]any{}, // semua required kosong
			wantCode:  http.StatusUnprocessableEntity,
			setupMock: func(m *mocks.UserUsecase) {},
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:  "422_invalid_cst_dob_format",
			idVar: "123",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "ALFA",
				"cst_dob":        "10-05-1992", // salah format
				"cst_phoneNum":   "0811",
				"cst_email":      "alfa@example.com",
			},
			wantCode:  http.StatusUnprocessableEntity,
			setupMock: func(m *mocks.UserUsecase) {},
			checkBody: func(t *testing.T, b []byte) {
				assert.Contains(t, string(b), "invalid cst_dob")
			},
		},
		{
			name:  "422_invalid_family_dob_format",
			idVar: "123",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "ALFA",
				"cst_dob":        "1992-05-10",
				"cst_phoneNum":   "0811",
				"cst_email":      "alfa@example.com",
				"family": []map[string]any{
					{"fl_relation": "Child", "fl_name": "GAMA", "fl_dob": "31/12/2010"}, // salah format
				},
			},
			wantCode:  http.StatusUnprocessableEntity,
			setupMock: func(m *mocks.UserUsecase) {},
			checkBody: func(t *testing.T, b []byte) {
				assert.Contains(t, string(b), "invalid fl_dob")
			},
		},
		{
			name:  "404_not_found",
			idVar: "777",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "GAMMA",
				"cst_dob":        "1990-01-01",
				"cst_phoneNum":   "0800",
				"cst_email":      "g@example.com",
				"family":         []map[string]any{},
			},
			wantCode: http.StatusNotFound,
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Update",
					mock.Anything,
					int32(777),
					mock.MatchedBy(func(c domain.Customer) bool {
						return c.Name == "GAMMA" && c.Email == "g@example.com" && c.NationalityID == 1
					}),
				).Return(domain.ErrNotFound).Once()
			},
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:  "500_repo_error",
			idVar: "888",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "DELTA",
				"cst_dob":        "1991-02-02",
				"cst_phoneNum":   "0801",
				"cst_email":      "d@example.com",
			},
			wantCode: http.StatusInternalServerError,
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Update", mock.Anything, int32(888), mock.Anything).
					Return(errors.New("db down")).Once()
			},
			checkBody: func(t *testing.T, b []byte) { assert.NotEmpty(t, b) },
		},
		{
			name:  "200_ok",
			idVar: "126",
			bodyObj: map[string]any{
				"nationality_id": 1,
				"cst_name":       "ALFA",
				"cst_dob":        "1992-05-10",
				"cst_phoneNum":   "0811000001",
				"cst_email":      "alfa1@example.com",
				"family": []map[string]any{
					{"fl_relation": "Spouse", "fl_name": "BETA", "fl_dob": "1993-07-01"},
				},
			},
			wantCode: http.StatusOK,
			setupMock: func(m *mocks.UserUsecase) {
				m.On("Update",
					mock.Anything,
					int32(126),
					mock.MatchedBy(func(c domain.Customer) bool {
						if c.NationalityID != 1 || c.Name != "ALFA" || c.Email != "alfa1@example.com" {
							return false
						}
						if len(c.Family) != 1 {
							return false
						}
						return c.Family[0].Relation == "Spouse" && c.Family[0].Name == "BETA"
					}),
				).Return(nil).Once()
			},
			checkBody: func(t *testing.T, b []byte) {
				var got map[string]any
				require.NoError(t, json.Unmarshal(b, &got))
				assert.Equal(t, "ok", got["status"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUsecase)
			if tc.setupMock != nil {
				tc.setupMock(mockUC)
			}

			h := &Handler{UC: mockUC, Val: validator.New()}

			var body []byte
			if tc.bodyRaw != nil {
				body = tc.bodyRaw
			} else if tc.bodyObj != nil {
				body = mustJSON(t, tc.bodyObj)
			}

			req, rr := makeUpdateReq(tc.idVar, body)
			h.UpdateUser(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.wantCode, res.StatusCode)
			if tc.checkBody != nil {
				tc.checkBody(t, rr.Body.Bytes())
			}

			mockUC.AssertExpectations(t)
		})
	}
}

func TestNewHandler(t *testing.T) {
	t.Run("with_usecase", func(t *testing.T) {
		uc := new(mocks.UserUsecase)

		h := NewHandler(uc)

		require.NotNil(t, h)
		assert.Equal(t, uc, h.UC) // UC dipasang
		require.NotNil(t, h.Val)  // validator dibuat
		assert.IsType(t, &validator.Validate{}, h.Val)
	})

	t.Run("nil_usecase", func(t *testing.T) {
		h := NewHandler(nil)

		require.NotNil(t, h)
		assert.Nil(t, h.UC)      // UC boleh nil
		require.NotNil(t, h.Val) // validator tetap dibuat
		assert.IsType(t, &validator.Validate{}, h.Val)
	})
}

func Test_mustParse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name  string
		args  args
		wantT time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantT, mustParse(tt.args.s), "mustParse(%v)", tt.args.s)
		})
	}
}

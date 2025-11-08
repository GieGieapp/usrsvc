package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"

	"usrsvc/internal/domain"
)

type PgUserRepo struct{ db *pgxpool.Pool }

func NewPgUserRepo(db *pgxpool.Pool) *PgUserRepo { return &PgUserRepo{db: db} }

func (r *PgUserRepo) ListCustomers(ctx context.Context, search string, limit, offset int) ([]domain.Customer, int32, error) {

	q := `SELECT cst_id, nationality_id, cst_name, cst_dob, cst_phoneNum, cst_email
	      FROM customer WHERE ($1='' OR cst_name ILIKE '%'||$1||'%' OR cst_email ILIKE '%'||$1||'%')
	      ORDER BY cst_id DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, strings.TrimSpace(search), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.Customer
	for rows.Next() {
		var c domain.Customer
		if err := rows.Scan(&c.ID, &c.NationalityID, &c.Name, &c.Dob, &c.PhoneNum, &c.Email); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	var total int32
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM customer WHERE ($1='' OR cst_name ILIKE '%'||$1||'%' OR cst_email ILIKE '%'||$1||'%')`, strings.TrimSpace(search)).Scan(&total)
	return out, total, nil
}

func (r *PgUserRepo) GetCustomer(ctx context.Context, id int32) (*domain.Customer, error) {

	var c domain.Customer
	err := r.db.QueryRow(ctx, `SELECT cst_id,nationality_id,cst_name,cst_dob,cst_phoneNum,cst_email FROM customer WHERE cst_id=$1`, id).
		Scan(&c.ID, &c.NationalityID, &c.Name, &c.Dob, &c.PhoneNum, &c.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT fl_id,cst_id,fl_relation,fl_name,fl_dob FROM family_list WHERE cst_id=$1 ORDER BY fl_id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var f domain.FamilyMember
		if err := rows.Scan(&f.ID, &f.CustomerID, &f.Relation, &f.Name, &f.Dob); err != nil {
			return nil, err
		}
		c.Family = append(c.Family, f)
	}
	return &c, nil
}

func (r *PgUserRepo) CreateCustomer(ctx context.Context, c domain.Customer) (int32, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var id int32
	if err := tx.QueryRow(ctx,
		`INSERT INTO customer (nationality_id,cst_name,cst_dob,cst_phoneNum,cst_email)
		 VALUES ($1,$2,$3,$4,$5) RETURNING cst_id`,
		c.NationalityID, c.Name, c.Dob, c.PhoneNum, c.Email,
	).Scan(&id); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return 0, domain.ErrConflict
		}
		return 0, err
	}

	for _, f := range c.Family {
		if _, err := tx.Exec(ctx,
			`INSERT INTO family_list (cst_id,fl_relation,fl_name,fl_dob)
			 VALUES ($1,$2,$3,$4)`,
			id, f.Relation, f.Name, f.Dob,
		); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PgUserRepo) UpdateCustomer(ctx context.Context, id int32, c domain.Customer) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`UPDATE customer SET nationality_id=$1,cst_name=$2,cst_dob=$3,cst_phoneNum=$4,cst_email=$5 WHERE cst_id=$6`,
		c.NationalityID, c.Name, c.Dob, c.PhoneNum, c.Email, id); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM family_list WHERE cst_id=$1`, id); err != nil {
		return err
	}
	for _, f := range c.Family {
		if _, err := tx.Exec(ctx,
			`INSERT INTO family_list (cst_id,fl_relation,fl_name,fl_dob) VALUES ($1,$2,$3,$4)`,
			id, f.Relation, f.Name, f.Dob); err != nil {
			return err
		}
	}
	return nil
}

func (r *PgUserRepo) DeleteCustomer(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM customer WHERE cst_id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PgUserRepo) ListNationalities(ctx context.Context) ([]domain.Nationality, error) {
	rows, err := r.db.Query(context.Background(),
		`SELECT nationality_id,nationality_name,nationality_code FROM nationality ORDER BY nationality_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Nationality
	for rows.Next() {
		var n domain.Nationality
		if err := rows.Scan(&n.ID, &n.Name, &n.Code); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

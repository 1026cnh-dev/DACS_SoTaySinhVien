package repository

import (
	"database/sql"

	"sotaysinhvien/internal/model"
)

const userSelectColumns = `id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at`

func scanUserRows(rows *sql.Rows) (map[int]model.User, error) {
	out := map[int]model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt); err != nil {
			return nil, err
		}
		out[u.ID] = u
	}
	return out, rows.Err()
}

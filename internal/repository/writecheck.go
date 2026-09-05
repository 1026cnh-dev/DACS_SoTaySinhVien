package repository

import (
	"database/sql"
	"fmt"
)

func requireAffected(result sql.Result, entity string) error {
	if result == nil {
		return fmt.Errorf("%s: không nhận được kết quả ghi dữ liệu", entity)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: không thể xác nhận dữ liệu đã ghi: %w", entity, err)
	}
	if n == 0 {
		return fmt.Errorf("%s không tồn tại hoặc không còn hợp lệ", entity)
	}
	return nil
}

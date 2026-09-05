package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"sotaysinhvien/internal/model"
	"sotaysinhvien/internal/repository"
)

type filterCacheEntry struct {
	items   []model.FilterValue
	expires time.Time
}

type postCacheEntry struct {
	items   []model.Post
	expires time.Time
}

type blockedCacheEntry struct {
	ids     []int
	expires time.Time
}

type ContentService struct {
	repo repository.ContentRepository

	cacheMu         sync.RWMutex
	categoriesCache []model.Category
	categoriesUntil time.Time
	activeAdsCache  []model.Advertisement
	activeAdsUntil  time.Time
	filterCache     map[string]filterCacheEntry
	feedCache       map[string]postCacheEntry
	blockedCache    map[int]blockedCacheEntry
}

const (
	categoriesCacheTTL = 60 * time.Second
	adsCacheTTL        = 30 * time.Second
	filterCacheTTL     = 45 * time.Second
	pinnedCacheTTL     = 10 * time.Second
	latestCacheTTL     = 6 * time.Second
	todayCacheTTL      = 6 * time.Second
	trendingCacheTTL   = 6 * time.Second
	blockedCacheTTL    = 20 * time.Second
)

func NewContentService(repo repository.ContentRepository) *ContentService {
	return &ContentService{repo: repo, filterCache: map[string]filterCacheEntry{}, feedCache: map[string]postCacheEntry{}, blockedCache: map[int]blockedCacheEntry{}}
}

func cloneCategories(items []model.Category) []model.Category {
	out := make([]model.Category, len(items))
	copy(out, items)
	return out
}
func cloneAds(items []model.Advertisement) []model.Advertisement {
	out := make([]model.Advertisement, len(items))
	copy(out, items)
	return out
}
func cloneFilterValues(items []model.FilterValue) []model.FilterValue {
	out := make([]model.FilterValue, len(items))
	copy(out, items)
	return out
}
func clonePosts(items []model.Post) []model.Post {
	out := make([]model.Post, len(items))
	copy(out, items)
	return out
}
func (s *ContentService) invalidateCategories() {
	s.cacheMu.Lock()
	s.categoriesCache = nil
	s.categoriesUntil = time.Time{}
	s.cacheMu.Unlock()
}
func (s *ContentService) invalidateAds() {
	s.cacheMu.Lock()
	s.activeAdsCache = nil
	s.activeAdsUntil = time.Time{}
	s.cacheMu.Unlock()
}
func (s *ContentService) invalidateFilters() {
	s.cacheMu.Lock()
	s.filterCache = map[string]filterCacheEntry{}
	s.cacheMu.Unlock()
}
func (s *ContentService) invalidateFeed() {
	s.cacheMu.Lock()
	s.feedCache = map[string]postCacheEntry{}
	s.cacheMu.Unlock()
}
func (s *ContentService) cachedPosts(key string, ttl time.Duration, loader func() ([]model.Post, error)) ([]model.Post, error) {
	now := time.Now()
	s.cacheMu.RLock()
	entry, ok := s.feedCache[key]
	if ok && now.Before(entry.expires) {
		items := clonePosts(entry.items)
		s.cacheMu.RUnlock()
		return items, nil
	}
	s.cacheMu.RUnlock()
	items, err := loader()
	if err != nil {
		return nil, err
	}
	s.cacheMu.Lock()
	s.feedCache[key] = postCacheEntry{items: clonePosts(items), expires: now.Add(ttl)}
	s.cacheMu.Unlock()
	return items, nil
}

func FieldDefinitions() []model.FieldConfig {
	return []model.FieldConfig{
		{Key: "title", Label: "Tiêu đề", Enabled: true, Required: true, Order: 10},
		{Key: "content", Label: "Nội dung", Enabled: true, Required: true, Order: 20},
		{Key: "image", Label: "Ảnh bài viết", Enabled: true, Required: true, Order: 30},
		{Key: "deadline", Label: "Hạn cuối", Enabled: false, Required: false, Order: 40},
		{Key: "company_logo", Label: "Logo công ty", Enabled: false, Required: false, Order: 50},
		{Key: "website", Label: "Website", Enabled: false, Required: false, Order: 60},
		{Key: "fanpage", Label: "Fanpage", Enabled: false, Required: false, Order: 70},
		{Key: "recruitment_content", Label: "Nội dung tuyển dụng", Enabled: false, Required: false, Order: 80},
		{Key: "cv_email", Label: "Email nhận CV ứng tuyển", Enabled: false, Required: false, Order: 90},
		{Key: "positions", Label: "Các vị trí tuyển", Enabled: false, Required: false, Order: 100},
		{Key: "contact_name", Label: "Người liên hệ", Enabled: false, Required: false, Order: 110},
		{Key: "contact_phone", Label: "Số điện thoại liên hệ", Enabled: false, Required: false, Order: 120},
		{Key: "organization", Label: "Đơn vị / tổ chức", Enabled: false, Required: false, Order: 130},
		{Key: "location", Label: "Địa điểm", Enabled: false, Required: false, Order: 140},
		{Key: "salary_range", Label: "Mức lương / hỗ trợ", Enabled: false, Required: false, Order: 150},
		{Key: "application_link", Label: "Liên kết đăng ký / ứng tuyển", Enabled: false, Required: false, Order: 160},
		{Key: "event_time", Label: "Thời gian diễn ra", Enabled: false, Required: false, Order: 170},
		{Key: "audience", Label: "Đối tượng tham gia", Enabled: false, Required: false, Order: 180},
		{Key: "tags", Label: "Từ khóa / thẻ", Enabled: false, Required: false, Order: 190},
		{Key: "source", Label: "Nguồn thông tin", Enabled: false, Required: false, Order: 200},
		{Key: "document_file", Label: "File tài liệu học tập", Enabled: false, Required: false, Order: 210},
		{Key: "school", Label: "Trường học", Enabled: false, Required: false, Order: 220},
		{Key: "document_type", Label: "Loại tài liệu", Enabled: false, Required: false, Order: 230},
		{Key: "subject", Label: "Môn học", Enabled: false, Required: false, Order: 240},
		{Key: "academic_year", Label: "Niên khóa", Enabled: false, Required: false, Order: 250},
	}
}

func DefaultFieldConfig(slug string) []model.FieldConfig {
	defs := FieldDefinitions()
	if slug == "hoc-tap" {
		for i := range defs {
			if map[string]bool{"title": true, "content": true, "image": true, "document_file": true, "school": true, "document_type": true, "subject": true, "academic_year": true, "tags": true, "source": true}[defs[i].Key] {
				defs[i].Enabled = true
			}
			if map[string]bool{"title": true, "content": true, "document_file": true, "school": true, "document_type": true, "subject": true, "academic_year": true}[defs[i].Key] {
				defs[i].Required = true
			}
			if map[string]bool{"school": true, "document_type": true, "subject": true, "academic_year": true}[defs[i].Key] {
				defs[i].Filterable = true
				defs[i].SuggestValues = true
				defs[i].AllowCustom = true
			}
		}
	}
	if slug == "viec-lam" {
		for i := range defs {
			defs[i].Enabled = true
			if defs[i].Key == "deadline" || defs[i].Key == "fanpage" || defs[i].Key == "recruitment_content" || defs[i].Key == "cv_email" || defs[i].Key == "positions" {
				defs[i].Required = true
			}
		}
	}
	return defs
}

func normalizeFieldLabel(raw, fallback string) string {
	label := strings.TrimSpace(raw)
	if label == "" {
		return fallback
	}
	runes := []rune(label)
	if len(runes) > 80 {
		label = string(runes[:80])
	}
	return label
}

func normalizeFieldType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "textarea":
		return "textarea"
	case "number":
		return "number"
	case "email":
		return "email"
	case "url":
		return "url"
	case "date":
		return "date"
	default:
		return "text"
	}
}

func isCustomFieldKey(key string) bool {
	key = strings.TrimSpace(key)
	if !strings.HasPrefix(key, "custom_") || len(key) < 8 || len(key) > 80 {
		return false
	}
	for _, ch := range key {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			continue
		}
		return false
	}
	return true
}

func NormalizeConfig(slug, raw string) string {
	defs := FieldDefinitions()
	defaults := DefaultFieldConfig(slug)
	if strings.TrimSpace(raw) == "" || strings.TrimSpace(raw) == "[]" {
		b, _ := json.Marshal(defaults)
		return string(b)
	}
	legacyFilterConfig := !strings.Contains(raw, `"filterable"`)
	var cfg []model.FieldConfig
	if json.Unmarshal([]byte(raw), &cfg) != nil {
		b, _ := json.Marshal(defaults)
		return string(b)
	}
	byKey := make(map[string]model.FieldConfig, len(cfg))
	maxOrder := 0
	for _, item := range cfg {
		item.Key = strings.TrimSpace(item.Key)
		if item.Key == "" {
			continue
		}
		byKey[item.Key] = item
		if item.Order > maxOrder {
			maxOrder = item.Order
		}
	}
	out := make([]model.FieldConfig, 0, len(defs)+8)
	standardKeys := make(map[string]bool, len(defs))
	for i, def := range defs {
		standardKeys[def.Key] = true
		if cur, ok := byKey[def.Key]; ok {
			cur.Label = normalizeFieldLabel(cur.Label, def.Label)
			if legacyFilterConfig {
				cur.Filterable = defaults[i].Filterable
				cur.SuggestValues = defaults[i].SuggestValues
				cur.AllowCustom = defaults[i].AllowCustom
			}
			if cur.Order <= 0 {
				cur.Order = def.Order
			}
			out = append(out, cur)
		} else {
			def.Enabled = defaults[i].Enabled
			def.Required = defaults[i].Required
			if maxOrder > 0 {
				maxOrder += 10
				def.Order = maxOrder
			}
			out = append(out, def)
		}
	}
	for _, item := range cfg {
		if standardKeys[item.Key] || !isCustomFieldKey(item.Key) {
			continue
		}
		item.Label = normalizeFieldLabel(item.Label, "Trường tùy chỉnh")
		item.Type = normalizeFieldType(item.Type)
		if item.Order <= 0 {
			maxOrder += 10
			item.Order = maxOrder
		}
		if item.Required {
			item.Enabled = true
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	b, _ := json.Marshal(out)
	return string(b)
}

func normalizeCardStyle(style string) string {
	switch strings.TrimSpace(strings.ToLower(style)) {
	case "horizontal":
		return "horizontal"
	case "vertical":
		return "vertical"
	default:
		return "full"
	}
}

func normalizePostCardStyle(style string) string {
	switch strings.TrimSpace(strings.ToLower(style)) {
	case "horizontal":
		return "horizontal"
	case "document":
		return "document"
	default:
		return "normal"
	}
}

func NormalizeDocumentFormats(raw string) string {
	allowed := map[string]bool{"pdf": true, "doc": true, "docx": true, "xls": true, "xlsx": true, "ppt": true, "pptx": true, "txt": true, "zip": true}
	seen := map[string]bool{}
	out := make([]string, 0, 9)
	for _, part := range strings.Split(strings.ToLower(raw), ",") {
		v := strings.TrimSpace(strings.TrimPrefix(part, "."))
		if allowed[v] && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return "pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip"
	}
	return strings.Join(out, ",")
}

func normalizeAudienceScope(scope string) string {
	allowed := map[string]bool{"public": true, "members": true, "students": true, "same_school": true, "admin": true}
	seen := map[string]bool{}
	ordered := make([]string, 0, 5)
	for _, part := range strings.Split(scope, ",") {
		v := strings.ToLower(strings.TrimSpace(part))
		if !allowed[v] || seen[v] {
			continue
		}
		seen[v] = true
		ordered = append(ordered, v)
	}
	if len(ordered) == 0 || seen["public"] {
		return "public"
	}
	return strings.Join(ordered, ",")
}

func (s *ContentService) ListCategories() ([]model.Category, error) {
	now := time.Now()
	s.cacheMu.RLock()
	if len(s.categoriesCache) > 0 && now.Before(s.categoriesUntil) {
		cats := cloneCategories(s.categoriesCache)
		s.cacheMu.RUnlock()
		return cats, nil
	}
	s.cacheMu.RUnlock()
	cats, err := s.repo.ListCategories()
	if err != nil {
		return nil, err
	}
	for i := range cats {
		cats[i].FormConfig = NormalizeConfig(cats[i].Slug, cats[i].FormConfig)
		cats[i].CompanyCardStyle = normalizeCardStyle(cats[i].CompanyCardStyle)
		cats[i].AudienceScope = normalizeAudienceScope(cats[i].AudienceScope)
		cats[i].PostCardStyle = normalizePostCardStyle(cats[i].PostCardStyle)
		cats[i].DocumentFormats = NormalizeDocumentFormats(cats[i].DocumentFormats)
	}
	s.cacheMu.Lock()
	s.categoriesCache = cloneCategories(cats)
	s.categoriesUntil = now.Add(categoriesCacheTTL)
	s.cacheMu.Unlock()
	return cats, nil
}
func (s *ContentService) ListPosts(q, category string) ([]model.Post, error) {
	q = strings.TrimSpace(q)
	category = strings.TrimSpace(category)
	// Trang chủ cần toàn bộ bài theo thời gian mới nhất. Cache ngắn giúp giữ tốc độ
	// với Supabase nhưng vẫn làm bài vừa đăng xuất hiện gần như ngay lập tức.
	if q == "" {
		key := fmt.Sprintf("latest|%s", category)
		return s.cachedPosts(key, latestCacheTTL, func() ([]model.Post, error) { return s.repo.ListPosts(q, category) })
	}
	return s.repo.ListPosts(q, category)
}
func (s *ContentService) ListPostsLimited(q, category string, limit int) ([]model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	return s.repo.ListPostsLimited(strings.TrimSpace(q), strings.TrimSpace(category), limit)
}
func (s *ContentService) ListPostsByAuthor(authorID int) ([]model.Post, error) {
	if authorID <= 0 {
		return nil, errors.New("người dùng không hợp lệ")
	}
	return s.repo.ListPostsByAuthor(authorID)
}
func (s *ContentService) ListPinnedPosts(category string, limit int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 4
	}
	category = strings.TrimSpace(category)
	key := fmt.Sprintf("pinned|%s|%d", category, limit)
	return s.cachedPosts(key, pinnedCacheTTL, func() ([]model.Post, error) { return s.repo.ListPinnedPosts(category, limit) })
}
func (s *ContentService) ListTodayPosts(category string, limit int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 8
	}
	category = strings.TrimSpace(category)
	key := fmt.Sprintf("today|%s|%d", category, limit)
	return s.cachedPosts(key, todayCacheTTL, func() ([]model.Post, error) { return s.repo.ListTodayPosts(category, limit) })
}
func (s *ContentService) SetPostPinned(id int, pinned bool) error {
	if id <= 0 {
		return errors.New("id không hợp lệ")
	}
	err := s.repo.SetPostPinned(id, pinned)
	if err == nil {
		s.invalidateFeed()
	}
	return err
}
func (s *ContentService) GetPost(id int) (model.Post, error) { return s.repo.GetPost(id) }
func (s *ContentService) SavePost(p model.Post) error {
	p.Title = strings.TrimSpace(p.Title)
	p.Content = strings.TrimSpace(p.Content)
	p.Summary = strings.TrimSpace(p.Summary)
	if p.Title == "" || p.Content == "" || p.CategoryID == 0 {
		return errors.New("vui lòng nhập đủ dữ liệu bắt buộc")
	}
	if len([]rune(p.Title)) > 180 {
		return errors.New("tiêu đề tối đa 180 ký tự")
	}
	if len([]rune(p.Content)) > 30000 {
		return errors.New("nội dung tối đa 30.000 ký tự")
	}
	if len([]rune(p.Summary)) > 600 {
		return errors.New("mô tả ngắn tối đa 600 ký tự")
	}
	if strings.TrimSpace(p.MetaJSON) == "" {
		p.MetaJSON = "{}"
	}
	err := s.repo.SavePost(p)
	if err == nil {
		s.invalidateFeed()
	}
	return err
}
func (s *ContentService) DeletePost(id int) error {
	if id <= 0 {
		return errors.New("id không hợp lệ")
	}
	err := s.repo.DeletePost(id)
	if err == nil {
		s.invalidateFeed()
	}
	return err
}
func (s *ContentService) SaveCategory(c model.Category) error {
	c.Name = strings.TrimSpace(c.Name)
	c.Slug = strings.TrimSpace(c.Slug)
	if c.Name == "" || c.Slug == "" {
		return errors.New("thiếu tên hoặc slug")
	}
	c.FormConfig = NormalizeConfig(c.Slug, c.FormConfig)
	c.CompanyCardStyle = normalizeCardStyle(c.CompanyCardStyle)
	c.AudienceScope = normalizeAudienceScope(c.AudienceScope)
	c.PostCardStyle = normalizePostCardStyle(c.PostCardStyle)
	c.DocumentFormats = NormalizeDocumentFormats(c.DocumentFormats)
	err := s.repo.SaveCategory(c)
	if err == nil {
		s.invalidateCategories()
	}
	return err
}
func (s *ContentService) UpdateCategoryConfig(c model.Category) error {
	if c.ID <= 0 {
		return errors.New("id không hợp lệ")
	}
	c.FormConfig = NormalizeConfig(c.Slug, c.FormConfig)
	c.CompanyCardStyle = normalizeCardStyle(c.CompanyCardStyle)
	c.AudienceScope = normalizeAudienceScope(c.AudienceScope)
	c.PostCardStyle = normalizePostCardStyle(c.PostCardStyle)
	c.DocumentFormats = NormalizeDocumentFormats(c.DocumentFormats)
	err := s.repo.UpdateCategoryConfig(c)
	if err == nil {
		s.invalidateCategories()
	}
	return err
}
func (s *ContentService) UpdateCategoryMeta(c model.Category) error {
	if c.ID <= 0 {
		return errors.New("id không hợp lệ")
	}
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return errors.New("tên chuyên mục không được để trống")
	}
	if len([]rune(c.Name)) > 80 {
		return errors.New("tên chuyên mục tối đa 80 ký tự")
	}
	c.AudienceScope = normalizeAudienceScope(c.AudienceScope)
	c.PostCardStyle = normalizePostCardStyle(c.PostCardStyle)
	c.CompanyCardStyle = normalizeCardStyle(c.CompanyCardStyle)
	c.DocumentFormats = NormalizeDocumentFormats(c.DocumentFormats)
	err := s.repo.UpdateCategoryMeta(c)
	if err == nil {
		s.invalidateCategories()
	}
	return err
}
func (s *ContentService) ReorderCategories(ids []int) error {
	err := s.repo.ReorderCategories(ids)
	if err == nil {
		s.invalidateCategories()
	}
	return err
}
func (s *ContentService) DeleteCategory(id int) error {
	err := s.repo.DeleteCategory(id)
	if err == nil {
		s.invalidateCategories()
		s.invalidateFilters()
	}
	return err
}

func BuildFieldConfig(values map[string][]string) string {
	defs := FieldDefinitions()
	for i := range defs {
		key := defs[i].Key
		if raw := values["label_"+key]; len(raw) > 0 {
			defs[i].Label = normalizeFieldLabel(raw[0], defs[i].Label)
		}
		defs[i].Enabled = len(values["enabled_"+key]) > 0
		defs[i].Required = len(values["required_"+key]) > 0
		defs[i].Filterable = len(values["filterable_"+key]) > 0
		defs[i].SuggestValues = len(values["suggest_"+key]) > 0
		defs[i].AllowCustom = len(values["allow_custom_"+key]) > 0
		if defs[i].Required {
			defs[i].Enabled = true
		}
		if raw := values["order_"+key]; len(raw) > 0 {
			var n int
			for _, ch := range raw[0] {
				if ch >= '0' && ch <= '9' {
					n = n*10 + int(ch-'0')
				}
			}
			if n > 0 {
				defs[i].Order = n
			}
		}
	}
	seen := map[string]bool{}
	for _, key := range values["custom_field_key"] {
		key = strings.TrimSpace(key)
		if seen[key] || !isCustomFieldKey(key) {
			continue
		}
		seen[key] = true
		label := "Trường tùy chỉnh"
		if raw := values["label_"+key]; len(raw) > 0 {
			label = normalizeFieldLabel(raw[0], label)
		}
		typeName := "text"
		if raw := values["type_"+key]; len(raw) > 0 {
			typeName = normalizeFieldType(raw[0])
		}
		item := model.FieldConfig{Key: key, Label: label, Type: typeName, Enabled: len(values["enabled_"+key]) > 0, Required: len(values["required_"+key]) > 0, Filterable: len(values["filterable_"+key]) > 0, SuggestValues: len(values["suggest_"+key]) > 0, AllowCustom: len(values["allow_custom_"+key]) > 0, Order: 9999}
		if item.Required {
			item.Enabled = true
		}
		if raw := values["order_"+key]; len(raw) > 0 {
			var n int
			for _, ch := range raw[0] {
				if ch >= '0' && ch <= '9' {
					n = n*10 + int(ch-'0')
				}
			}
			if n > 0 {
				item.Order = n
			}
		}
		defs = append(defs, item)
	}
	sort.SliceStable(defs, func(i, j int) bool { return defs[i].Order < defs[j].Order })
	b, _ := json.Marshal(defs)
	return string(b)
}

func (s *ContentService) UpsertFilterValue(categoryID int, fieldKey, value string, approved bool) error {
	value = strings.TrimSpace(value)
	if categoryID <= 0 || strings.TrimSpace(fieldKey) == "" || value == "" {
		return nil
	}
	err := s.repo.UpsertFilterValue(categoryID, strings.TrimSpace(fieldKey), value, approved)
	if err == nil {
		s.invalidateFilters()
	}
	return err
}
func (s *ContentService) ListApprovedFilterValuesForCategory(categoryID int) (map[string][]model.FilterValue, error) {
	key := fmt.Sprintf("%d|*", categoryID)
	now := time.Now()
	s.cacheMu.RLock()
	if entry, ok := s.filterCache[key]; ok && now.Before(entry.expires) {
		items := cloneFilterValues(entry.items)
		s.cacheMu.RUnlock()
		out := map[string][]model.FilterValue{}
		for _, item := range items {
			out[item.FieldKey] = append(out[item.FieldKey], item)
		}
		return out, nil
	}
	s.cacheMu.RUnlock()
	items, err := s.repo.ListFilterValues(categoryID, "", "approved")
	if err != nil {
		return nil, err
	}
	s.cacheMu.Lock()
	s.filterCache[key] = filterCacheEntry{items: cloneFilterValues(items), expires: now.Add(filterCacheTTL)}
	s.cacheMu.Unlock()
	out := map[string][]model.FilterValue{}
	for _, item := range items {
		out[item.FieldKey] = append(out[item.FieldKey], item)
	}
	return out, nil
}

func (s *ContentService) ListApprovedFilterValues(categoryID int, fieldKey string) ([]model.FilterValue, error) {
	fieldKey = strings.TrimSpace(fieldKey)
	key := fmt.Sprintf("%d|%s", categoryID, fieldKey)
	now := time.Now()
	s.cacheMu.RLock()
	if entry, ok := s.filterCache[key]; ok && now.Before(entry.expires) {
		items := cloneFilterValues(entry.items)
		s.cacheMu.RUnlock()
		return items, nil
	}
	s.cacheMu.RUnlock()
	items, err := s.repo.ListFilterValues(categoryID, fieldKey, "approved")
	if err != nil {
		return nil, err
	}
	s.cacheMu.Lock()
	s.filterCache[key] = filterCacheEntry{items: cloneFilterValues(items), expires: now.Add(filterCacheTTL)}
	s.cacheMu.Unlock()
	return items, nil
}
func (s *ContentService) ListFilterValues(categoryID int, fieldKey, status string) ([]model.FilterValue, error) {
	return s.repo.ListFilterValues(categoryID, strings.TrimSpace(fieldKey), strings.TrimSpace(status))
}
func (s *ContentService) ReviewFilterValue(id int, action, newValue string) error {
	err := s.repo.ReviewFilterValue(id, strings.TrimSpace(action), strings.TrimSpace(newValue))
	if err == nil {
		s.invalidateFilters()
	}
	return err
}

func (s *ContentService) ListTrendingPosts(limit int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 5
	}
	key := fmt.Sprintf("trending|%d", limit)
	return s.cachedPosts(key, trendingCacheTTL, func() ([]model.Post, error) { return s.repo.ListTrendingPosts(limit) })
}
func (s *ContentService) VotePost(postID, userID, value int) error {
	if postID <= 0 || userID <= 0 {
		return errors.New("dữ liệu bình chọn không hợp lệ")
	}
	if value != 1 && value != -1 {
		return errors.New("giá trị bình chọn không hợp lệ")
	}
	err := s.repo.VotePost(postID, userID, value)
	if err == nil {
		s.invalidateFeed()
	}
	return err
}
func (s *ContentService) GetPostStats(postID int) (model.PostStats, error) {
	return s.repo.GetPostStats(postID)
}
func (s *ContentService) GetPostStatsBatch(postIDs []int) (map[int]model.PostStats, error) {
	return s.repo.GetPostStatsBatch(postIDs)
}
func (s *ContentService) GetUserPostVote(postID, userID int) (int, error) {
	if userID <= 0 {
		return 0, nil
	}
	return s.repo.GetUserPostVote(postID, userID)
}
func (s *ContentService) AddComment(c model.Comment) error {
	c.Content = strings.TrimSpace(c.Content)
	if c.PostID <= 0 || c.UserID <= 0 || c.Content == "" {
		return errors.New("bình luận không hợp lệ")
	}
	return s.repo.CreateComment(c)
}
func (s *ContentService) ListComments(postID int) ([]model.Comment, error) {
	return s.repo.ListComments(postID)
}
func BuildCommentTree(items []model.Comment) []model.Comment {
	byParent := map[int][]model.Comment{}
	for _, item := range items {
		byParent[item.ParentID] = append(byParent[item.ParentID], item)
	}
	var walk func(int, int) []model.Comment
	walk = func(parent int, depth int) []model.Comment {
		children := byParent[parent]
		for i := range children {
			children[i].Depth = depth
			children[i].Children = walk(children[i].ID, depth+1)
		}
		return children
	}
	return walk(0, 0)
}

func (s *ContentService) ListAds(position string, activeOnly bool) ([]model.Advertisement, error) {
	position = strings.TrimSpace(position)
	if !activeOnly {
		return s.repo.ListAds(position, false)
	}
	now := time.Now()
	s.cacheMu.RLock()
	if s.activeAdsCache != nil && now.Before(s.activeAdsUntil) {
		all := cloneAds(s.activeAdsCache)
		s.cacheMu.RUnlock()
		if position == "" {
			return all, nil
		}
		out := make([]model.Advertisement, 0, len(all))
		for _, ad := range all {
			if ad.Position == position {
				out = append(out, ad)
			}
		}
		return out, nil
	}
	s.cacheMu.RUnlock()
	all, err := s.repo.ListAds("", true)
	if err != nil {
		return nil, err
	}
	s.cacheMu.Lock()
	s.activeAdsCache = cloneAds(all)
	s.activeAdsUntil = now.Add(adsCacheTTL)
	s.cacheMu.Unlock()
	if position == "" {
		return all, nil
	}
	out := make([]model.Advertisement, 0, len(all))
	for _, ad := range all {
		if ad.Position == position {
			out = append(out, ad)
		}
	}
	return out, nil
}
func (s *ContentService) GetAd(id int) (model.Advertisement, error) {
	if id <= 0 {
		return model.Advertisement{}, errors.New("id quảng cáo không hợp lệ")
	}
	return s.repo.GetAd(id)
}
func (s *ContentService) SaveAd(ad model.Advertisement) error {
	ad.Title = strings.TrimSpace(ad.Title)
	ad.Description = strings.TrimSpace(ad.Description)
	ad.LinkURL = strings.TrimSpace(ad.LinkURL)
	ad.Position = strings.TrimSpace(ad.Position)
	if ad.Position == "" {
		return errors.New("thiếu vị trí quảng cáo")
	}
	if ad.Title == "" && ad.Description == "" && strings.TrimSpace(ad.ImageURL) == "" {
		return errors.New("quảng cáo cần có ít nhất tiêu đề, nội dung hoặc hình ảnh")
	}
	if ad.SortOrder <= 0 {
		ad.SortOrder = 10
	}
	err := s.repo.SaveAd(ad)
	if err == nil {
		s.invalidateAds()
	}
	return err
}
func (s *ContentService) ReorderAds(ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	err := s.repo.ReorderAds(ids)
	if err == nil {
		s.invalidateAds()
	}
	return err
}

func (s *ContentService) DeleteAd(id int) error {
	if id <= 0 {
		return errors.New("id quảng cáo không hợp lệ")
	}
	err := s.repo.DeleteAd(id)
	if err == nil {
		s.invalidateAds()
	}
	return err
}

func (s *ContentService) SaveDraft(d model.Draft) (int, error) {
	d.Title = strings.TrimSpace(d.Title)
	d.Content = strings.TrimSpace(d.Content)
	if d.AuthorID <= 0 {
		return 0, errors.New("người dùng không hợp lệ")
	}
	if strings.TrimSpace(d.MetaJSON) == "" {
		d.MetaJSON = "{}"
	}
	return s.repo.SaveDraft(d)
}
func (s *ContentService) GetDraft(id, authorID int) (model.Draft, error) {
	return s.repo.GetDraft(id, authorID)
}
func (s *ContentService) ListDrafts(authorID int) ([]model.Draft, error) {
	return s.repo.ListDrafts(authorID)
}
func (s *ContentService) DeleteDraft(id, authorID int) error { return s.repo.DeleteDraft(id, authorID) }

func (s *ContentService) ListSavedPostIDs(userID int, postIDs []int) (map[int]bool, error) {
	if userID <= 0 || len(postIDs) == 0 {
		return map[int]bool{}, nil
	}
	return s.repo.ListSavedPostIDs(userID, postIDs)
}
func (s *ContentService) IsPostSaved(postID, userID int) (bool, error) {
	if postID <= 0 || userID <= 0 {
		return false, nil
	}
	return s.repo.IsPostSaved(postID, userID)
}
func (s *ContentService) ToggleSavedPost(postID, userID int) (bool, error) {
	if postID <= 0 || userID <= 0 {
		return false, errors.New("dữ liệu lưu bài không hợp lệ")
	}
	saved, err := s.repo.IsPostSaved(postID, userID)
	if err != nil {
		return false, err
	}
	if saved {
		if err := s.repo.UnsavePostForUser(postID, userID); err != nil {
			return false, err
		}
		return false, nil
	}
	if err := s.repo.SavePostForUser(postID, userID); err != nil {
		return false, err
	}
	return true, nil
}
func (s *ContentService) ListSavedPosts(userID int) ([]model.Post, error) {
	if userID <= 0 {
		return nil, errors.New("người dùng không hợp lệ")
	}
	return s.repo.ListSavedPosts(userID)
}

func (s *ContentService) ReportPost(postID, reporterID int, reason string) error {
	reason = strings.TrimSpace(reason)
	if postID <= 0 || reporterID <= 0 {
		return errors.New("dữ liệu báo cáo không hợp lệ")
	}
	if reason == "" {
		reason = "Nội dung không phù hợp"
	}
	return s.repo.CreatePostReport(model.PostReport{PostID: postID, ReporterID: reporterID, Reason: reason})
}
func (s *ContentService) ListPostReports(status string) ([]model.PostReport, error) {
	return s.repo.ListPostReports(status)
}
func (s *ContentService) UpdatePostReportStatus(id int, status string) error {
	if status != "pending" && status != "resolved" && status != "dismissed" {
		return errors.New("trạng thái kiểm duyệt không hợp lệ")
	}
	return s.repo.UpdatePostReportStatus(id, status)
}
func (s *ContentService) BlockUser(blockerID, blockedID int) error {
	err := s.repo.BlockUser(blockerID, blockedID)
	if err == nil {
		s.cacheMu.Lock()
		delete(s.blockedCache, blockerID)
		s.cacheMu.Unlock()
	}
	return err
}
func (s *ContentService) IsUserBlocked(blockerID, blockedID int) (bool, error) {
	return s.repo.IsUserBlocked(blockerID, blockedID)
}
func (s *ContentService) ListBlockedUserIDs(blockerID int) ([]int, error) {
	if blockerID <= 0 {
		return nil, nil
	}
	now := time.Now()
	s.cacheMu.RLock()
	entry, ok := s.blockedCache[blockerID]
	if ok && now.Before(entry.expires) {
		ids := append([]int(nil), entry.ids...)
		s.cacheMu.RUnlock()
		return ids, nil
	}
	s.cacheMu.RUnlock()
	ids, err := s.repo.ListBlockedUserIDs(blockerID)
	if err != nil {
		return nil, err
	}
	s.cacheMu.Lock()
	s.blockedCache[blockerID] = blockedCacheEntry{ids: append([]int(nil), ids...), expires: now.Add(blockedCacheTTL)}
	s.cacheMu.Unlock()
	return ids, nil
}

func (s *ContentService) CreateVerificationRequest(req model.VerificationRequest) error {
	req.Type = strings.TrimSpace(req.Type)
	req.Info = strings.TrimSpace(req.Info)
	if req.UserID <= 0 || req.Info == "" {
		return errors.New("vui lòng nhập đầy đủ thông tin xác thực")
	}
	if req.Type != "student" && req.Type != "employer" && req.Type != "landlord" && req.Type != "phone" {
		return errors.New("loại xác thực không hợp lệ")
	}
	return s.repo.CreateVerificationRequest(req)
}
func (s *ContentService) GetLatestVerificationRequest(userID int) (model.VerificationRequest, error) {
	return s.repo.GetLatestVerificationRequest(userID)
}
func (s *ContentService) HasPendingVerificationRequest(userID int, requestType string) (bool, error) {
	return s.repo.HasPendingVerificationRequest(userID, requestType)
}

func (s *ContentService) ListVerificationRequests(status string) ([]model.VerificationRequest, error) {
	return s.repo.ListVerificationRequests(status)
}
func (s *ContentService) ResolveVerificationRequest(id int, status string) error {
	if id <= 0 {
		return errors.New("yêu cầu xác thực không hợp lệ")
	}
	if status != "approved" && status != "rejected" {
		return errors.New("trạng thái xác thực không hợp lệ")
	}
	return s.repo.ResolveVerificationRequest(id, status)
}

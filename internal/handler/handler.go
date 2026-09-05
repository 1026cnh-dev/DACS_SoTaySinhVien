package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"sotaysinhvien/internal/model"
	"sotaysinhvien/internal/reference"
	"sotaysinhvien/internal/service"
	"sotaysinhvien/internal/session"
)

type Handler struct {
	content   *service.ContentService
	auth      *service.AuthService
	sessions  *session.Manager
	templates *template.Template
}

func New(content *service.ContentService, auth *service.AuthService, sessions *session.Manager, templates *template.Template) *Handler {
	return &Handler{content: content, auth: auth, sessions: sessions, templates: templates}
}
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.Home)
	mux.HandleFunc("/bai-viet/", h.PostDetail)
	mux.HandleFunc("/tai-lieu/xem-truoc/", h.DocumentPreview)
	mux.HandleFunc("/tai-lieu/xem-truoc-tam", h.DocumentUploadPreview)
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/auth/google", h.GoogleAuth)
	mux.HandleFunc("/auth/phone/check", h.PhoneAccountCheck)
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/ho-so", h.Profile)
	mux.HandleFunc("/verification/request", h.RequestVerification)
	mux.HandleFunc("/phone-verification/request", h.RequestPhoneVerification)
	mux.HandleFunc("/nguoi-dung/", h.PublicProfile)
	mux.HandleFunc("/dang-bai", h.SubmitPost)
	mux.HandleFunc("/bai-viet/sua", h.EditPost)
	mux.HandleFunc("/post/delete-own", h.DeleteOwnPost)
	mux.HandleFunc("/bai-da-luu", h.MyDrafts)
	mux.HandleFunc("/draft/delete", h.DeleteDraft)
	mux.HandleFunc("/post/vote", h.VotePost)
	mux.HandleFunc("/post/save", h.ToggleSavePost)
	mux.HandleFunc("/comment/add", h.AddComment)
	mux.HandleFunc("/post/report", h.ReportPost)
	mux.HandleFunc("/user/block", h.BlockUser)
	mux.HandleFunc("/admin", h.Admin)
	mux.HandleFunc("/admin/post/save", h.SavePost)
	mux.HandleFunc("/admin/post/pin", h.PinPost)
	mux.HandleFunc("/admin/post/delete", h.DeletePost)
	mux.HandleFunc("/admin/category/save", h.SaveCategory)
	mux.HandleFunc("/admin/category/config", h.SaveCategoryConfig)
	mux.HandleFunc("/admin/category/meta", h.UpdateCategoryMeta)
	mux.HandleFunc("/admin/category/reorder", h.ReorderCategories)
	mux.HandleFunc("/admin/category/delete", h.DeleteCategory)
	mux.HandleFunc("/admin/ad/save", h.SaveAd)
	mux.HandleFunc("/admin/ad/reorder", h.ReorderAds)
	mux.HandleFunc("/admin/ad/delete", h.DeleteAd)
	mux.HandleFunc("/admin/report/action", h.AdminReportAction)
	mux.HandleFunc("/admin/verification/action", h.AdminVerificationAction)
	mux.HandleFunc("/admin/user/action", h.AdminUserAction)
	mux.HandleFunc("/admin/filter-value/action", h.AdminFilterValueAction)
	mux.HandleFunc("/filter-options", h.FilterOptions)
	mux.HandleFunc("/api/education-options", h.EducationOptions)
}
func (h *Handler) render(w http.ResponseWriter, name string, data model.PageData) {
	// Only pages that render the shared header or category navigation need category data.
	if templateUsesSharedNavigation(name) && len(data.Categories) == 0 {
		if cats, err := h.content.ListCategories(); err == nil {
			data.Categories = cats
		}
	}
	if templateUsesSharedNavigation(name) && len(data.Categories) > 0 {
		data.Categories = decorateCategoryNavigation(data.Categories, data.CurrentUser)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func templateUsesSharedNavigation(name string) bool {
	switch name {
	case "home.html", "post.html", "profile.html", "admin.html", "submit.html", "drafts.html":
		return true
	default:
		return false
	}
}
func (h *Handler) withHeaderSession(r *http.Request, data model.PageData) model.PageData {
	if u, ok := h.sessions.CurrentUser(r); ok {
		data.IsLoggedIn = true
		data.UserName = u.Name
		data.CurrentUser = &u
	}
	return data
}
func userRestricted(u model.User) bool {
	return strings.EqualFold(strings.TrimSpace(u.AccountStatus), "restricted")
}

func rejectRestricted(w http.ResponseWriter, r *http.Request, u model.User) bool {
	if !userRestricted(u) || u.IsAdmin {
		return false
	}
	if wantsJSON(r) {
		writeAdminJSON(w, http.StatusForbidden, map[string]any{"ok": false, "message": "Tài khoản đang bị hạn chế thao tác."})
	} else {
		http.Error(w, "Tài khoản đang bị hạn chế thao tác. Vui lòng liên hệ quản trị viên.", http.StatusForbidden)
	}
	return true
}

func isStudentViewer(u *model.User) bool {
	if u == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(u.ProfileRole), "student") || strings.EqualFold(strings.TrimSpace(u.VerificationType), "student") || strings.TrimSpace(u.StudentID) != ""
}

func audienceScopes(raw string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		v := strings.ToLower(strings.TrimSpace(part))
		if v != "" {
			out[v] = true
		}
	}
	if len(out) == 0 {
		out["public"] = true
	}
	return out
}

func categoryVisibleTo(cat model.Category, viewer *model.User) bool {
	if viewer != nil && viewer.IsAdmin {
		return true
	}
	scopes := audienceScopes(cat.AudienceScope)
	if scopes["public"] {
		return true
	}
	if scopes["members"] && viewer != nil {
		return true
	}
	if scopes["students"] && isStudentViewer(viewer) {
		return true
	}
	if scopes["same_school"] && isStudentViewer(viewer) && strings.TrimSpace(viewer.School) != "" {
		return true
	}
	return false
}

func decorateCategoryNavigation(cats []model.Category, viewer *model.User) []model.Category {
	decorated := make([]model.Category, len(cats))
	copy(decorated, cats)
	for i := range decorated {
		// Admin has every permission, so the navigation never shows a lock for admin.
		if viewer != nil && viewer.IsAdmin {
			decorated[i].NavLocked = false
			decorated[i].NavAccessHint = ""
			continue
		}
		if categoryVisibleTo(decorated[i], viewer) {
			continue
		}
		decorated[i].NavLocked = true
		if viewer == nil {
			decorated[i].NavAccessHint = "Đăng nhập để xem nội dung chuyên mục này"
		} else {
			decorated[i].NavAccessHint = audienceAccessMessage(decorated[i], viewer)
		}
	}
	return decorated
}

func audienceAccessMessage(cat model.Category, viewer *model.User) string {
	if viewer != nil && viewer.IsAdmin {
		return ""
	}
	scopes := audienceScopes(cat.AudienceScope)
	if scopes["public"] {
		return ""
	}
	if viewer == nil {
		return "Vui lòng đăng nhập để xem nội dung chuyên mục này."
	}
	if scopes["members"] {
		return ""
	}
	if scopes["students"] || scopes["same_school"] {
		if !isStudentViewer(viewer) {
			return "Chuyên mục này dành cho sinh viên. Hãy cập nhật loại hồ sơ thành Sinh viên và bổ sung thông tin sinh viên."
		}
		if scopes["same_school"] && strings.TrimSpace(viewer.School) == "" {
			return "Chuyên mục này yêu cầu thông tin trường học. Hãy cập nhật Trường / Đại học trong hồ sơ để tiếp tục."
		}
	}
	if scopes["admin"] {
		return "Chuyên mục này chỉ dành cho quản trị viên."
	}
	return "Hồ sơ hiện tại chưa đáp ứng điều kiện xem nội dung chuyên mục này. Hãy cập nhật hồ sơ để tiếp tục."
}

func (h *Handler) redirectAudienceAccess(w http.ResponseWriter, r *http.Request, cat model.Category, viewer *model.User, message string) {
	target := r.URL.RequestURI()
	if viewer == nil {
		http.Redirect(w, r, "/login?next="+url.QueryEscape(target), http.StatusSeeOther)
		return
	}
	if strings.TrimSpace(message) == "" {
		message = audienceAccessMessage(cat, viewer)
	}
	profileURL := "/ho-so?msg=" + url.QueryEscape(message)
	// Admin-only access cannot be satisfied by editing a normal profile, so do not
	// create an automatic return loop in that case.
	scopes := audienceScopes(cat.AudienceScope)
	adminOnly := scopes["admin"] && !scopes["students"] && !scopes["same_school"] && !scopes["members"] && !scopes["public"]
	if !adminOnly {
		profileURL += "&next=" + url.QueryEscape(target)
	}
	http.Redirect(w, r, profileURL, http.StatusSeeOther)
}

// filterAccessibleCategories is intentionally used only in write flows such as
// creating/editing a post. Read navigation must keep every category visible.
func filterAccessibleCategories(cats []model.Category, viewer *model.User) []model.Category {
	out := make([]model.Category, 0, len(cats))
	for _, cat := range cats {
		if categoryVisibleTo(cat, viewer) {
			out = append(out, cat)
		}
	}
	return out
}

func categoryByID(cats []model.Category, id int) (model.Category, bool) {
	for _, cat := range cats {
		if cat.ID == id {
			return cat, true
		}
	}
	return model.Category{}, false
}

func (h *Handler) postVisibleTo(post model.Post, cats []model.Category, viewer *model.User, author *model.User) bool {
	cat, ok := categoryByID(cats, post.CategoryID)
	if !ok {
		return false
	}
	if viewer != nil && viewer.IsAdmin {
		return true
	}
	scopes := audienceScopes(cat.AudienceScope)
	if scopes["public"] || (scopes["members"] && viewer != nil) || (scopes["students"] && isStudentViewer(viewer)) {
		return true
	}
	if scopes["same_school"] {
		if !isStudentViewer(viewer) || author == nil || !isStudentViewer(author) {
			return false
		}
		viewerSchool := strings.TrimSpace(viewer.School)
		authorSchool := strings.TrimSpace(author.School)
		return viewerSchool != "" && authorSchool != "" && reference.SameSchool(viewerSchool, authorSchool)
	}
	return false
}

func (h *Handler) userCanAccessPost(postID int, viewer *model.User) bool {
	post, err := h.content.GetPost(postID)
	if err != nil {
		return false
	}
	cats, _ := h.content.ListCategories()
	var author *model.User
	if post.AuthorID > 0 {
		if u, err := h.auth.GetUserByID(post.AuthorID); err == nil {
			author = &u
		}
	}
	return h.postVisibleTo(post, cats, viewer, author)
}

func metaFilterValue(meta model.PostMeta, key string) string {
	switch key {
	case "school":
		return strings.TrimSpace(meta.School)
	case "document_type":
		return strings.TrimSpace(meta.DocumentType)
	case "subject":
		return strings.TrimSpace(meta.Subject)
	case "academic_year":
		return strings.TrimSpace(meta.AcademicYear)
	case "location":
		return strings.TrimSpace(meta.Location)
	case "organization":
		return strings.TrimSpace(meta.Organization)
	case "salary_range":
		return strings.TrimSpace(meta.SalaryRange)
	case "audience":
		return strings.TrimSpace(meta.Audience)
	case "tags":
		return strings.TrimSpace(meta.Tags)
	case "source":
		return strings.TrimSpace(meta.Source)
	case "event_time":
		return strings.TrimSpace(meta.EventTime)
	case "deadline":
		return strings.TrimSpace(meta.Deadline)
	}
	if meta.CustomFields != nil {
		return strings.TrimSpace(meta.CustomFields[key])
	}
	return ""
}
func fieldLabelFromConfig(cat model.Category, key string) string {
	for _, f := range enabledConfig(cat.FormConfig) {
		if f.Key == key {
			return f.Label
		}
	}
	return key
}

func validatePostMeta(meta model.PostMeta) error {
	checkURL := func(label, value string) error {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil
		}
		u, err := url.ParseRequestURI(value)
		if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return fmt.Errorf("%s không hợp lệ; chỉ chấp nhận http/https", label)
		}
		if len([]rune(value)) > 1000 {
			return fmt.Errorf("%s quá dài", label)
		}
		return nil
	}
	for _, item := range []struct{ label, value string }{
		{"Website", meta.Website}, {"Fanpage", meta.Fanpage}, {"Liên kết ứng tuyển", meta.ApplicationLink}, {"Nguồn", meta.Source},
	} {
		if err := checkURL(item.label, item.value); err != nil {
			return err
		}
	}
	if v := strings.TrimSpace(meta.CVEmail); v != "" {
		addr, err := mail.ParseAddress(v)
		if err != nil || !strings.EqualFold(addr.Address, v) || len(v) > 254 {
			return errors.New("Email nhận CV không hợp lệ")
		}
	}
	if v := strings.TrimSpace(meta.ContactPhone); v != "" {
		digits := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, v)
		if len(digits) < 9 || len(digits) > 12 {
			return errors.New("Số điện thoại liên hệ không hợp lệ")
		}
	}
	for label, value := range map[string]string{
		"Nội dung tuyển dụng": meta.RecruitmentContent, "Đơn vị / tổ chức": meta.Organization, "Địa điểm": meta.Location,
		"Đối tượng": meta.Audience, "Từ khóa": meta.Tags, "Trường": meta.School, "Môn học": meta.Subject,
	} {
		if len([]rune(strings.TrimSpace(value))) > 5000 {
			return fmt.Errorf("%s quá dài", label)
		}
	}
	return nil
}

func (h *Handler) syncFilterValues(cat model.Category, meta model.PostMeta) {
	for _, f := range enabledConfig(cat.FormConfig) {
		if !f.Enabled || (!f.Filterable && !f.SuggestValues) {
			continue
		}
		v := metaFilterValue(meta, f.Key)
		if v == "" {
			continue
		}
		// Giá trị người dùng tự tạo vẫn được lưu, nhưng mặc định chờ quản trị viên duyệt.
		approved := false
		if existing, err := h.content.ListApprovedFilterValues(cat.ID, f.Key); err == nil {
			for _, x := range existing {
				if strings.EqualFold(strings.TrimSpace(x.Value), v) {
					approved = true
					break
				}
			}
		}
		_ = h.content.UpsertFilterValue(cat.ID, f.Key, v, approved)
	}
}

func (h *Handler) loadAuthors(groups ...[]model.Post) map[int]*model.User {
	ids := make([]int, 0, 32)
	seen := map[int]bool{}
	for _, group := range groups {
		for _, post := range group {
			if post.AuthorID > 0 && !seen[post.AuthorID] {
				seen[post.AuthorID] = true
				ids = append(ids, post.AuthorID)
			}
		}
	}
	loaded, err := h.auth.GetUsersByIDs(ids)
	if err != nil {
		return map[int]*model.User{}
	}
	out := make(map[int]*model.User, len(loaded))
	for id, item := range loaded {
		u := item
		out[id] = &u
	}
	return out
}

func postIDs(groups ...[]model.Post) []int {
	ids := make([]int, 0, 32)
	seen := map[int]bool{}
	for _, group := range groups {
		for _, post := range group {
			if post.ID > 0 && !seen[post.ID] {
				seen[post.ID] = true
				ids = append(ids, post.ID)
			}
		}
	}
	return ids
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	catSlug := strings.TrimSpace(r.URL.Query().Get("category"))
	var allCats []model.Category
	var sessionUser model.User
	var sessionOK bool
	var headWG sync.WaitGroup
	headWG.Add(2)
	go func() {
		defer headWG.Done()
		allCats, _ = h.content.ListCategories()
	}()
	go func() {
		defer headWG.Done()
		sessionUser, sessionOK = h.sessions.CurrentUser(r)
	}()
	headWG.Wait()

	var viewer *model.User
	currentUserID := 0
	if sessionOK {
		viewer = &sessionUser
		currentUserID = sessionUser.ID
	}
	visibleCats := allCats
	var blockedCategory *model.Category
	if catSlug != "" {
		found := false
		for i := range allCats {
			c := allCats[i]
			if c.Slug == catSlug {
				found = true
				if !categoryVisibleTo(c, viewer) {
					blockedCategory = &allCats[i]
				}
				break
			}
		}
		if !found {
			catSlug = ""
		}
	}

	cn := ""
	for _, c := range visibleCats {
		if c.Slug == catSlug {
			cn = c.Name
			break
		}
	}
	data := model.PageData{Title: "Sổ tay sinh viên", Categories: visibleCats, Query: q, Category: catSlug, CategoryName: cn}
	if blockedCategory != nil {
		message := audienceAccessMessage(*blockedCategory, viewer)
		data.CategoryAccessBlocked = true
		data.CategoryAccessMessage = message
		target := "/?category=" + url.QueryEscape(blockedCategory.Slug) + "#feed-4-cot"
		if viewer == nil {
			data.CategoryAccessURL = "/login?next=" + url.QueryEscape(target)
			data.CategoryAccessAction = "Đăng nhập để tiếp tục"
		} else {
			scopes := audienceScopes(blockedCategory.AudienceScope)
			adminOnly := scopes["admin"] && !scopes["students"] && !scopes["same_school"] && !scopes["members"] && !scopes["public"]
			if adminOnly {
				data.CategoryAccessURL = "/"
				data.CategoryAccessAction = "Quay lại trang chủ"
			} else {
				data.CategoryAccessURL = "/ho-so?msg=" + url.QueryEscape(message) + "&next=" + url.QueryEscape(target)
				data.CategoryAccessAction = "Cập nhật hồ sơ"
			}
		}
	}

	categoryStyle := make(map[int]string, len(allCats))
	var selectedCategory *model.Category
	for i := range allCats {
		categoryStyle[allCats[i].ID] = allCats[i].PostCardStyle
		if allCats[i].Slug == catSlug {
			selectedCategory = &allCats[i]
			if allCats[i].PostCardStyle == "document" {
				data.IsDocumentCategory = true
			}
		}
	}

	activeFilterValues := map[string]string{}
	if selectedCategory != nil && blockedCategory == nil {
		approvedByField, _ := h.content.ListApprovedFilterValuesForCategory(selectedCategory.ID)
		for _, f := range enabledConfig(selectedCategory.FormConfig) {
			if !f.Enabled || !f.Filterable {
				continue
			}
			selected := strings.TrimSpace(r.URL.Query().Get("f_" + f.Key))
			activeFilterValues[f.Key] = selected
			group := model.FilterGroup{Key: f.Key, Label: f.Label, Selected: selected}
			for _, v := range approvedByField[f.Key] {
				group.Options = append(group.Options, model.FilterOption{Value: v.Value, Count: v.UsageCount})
			}
			data.FilterGroups = append(data.FilterGroups, group)
		}
	}
	if viewer != nil {
		data.IsLoggedIn = true
		data.UserName = viewer.Name
		data.CurrentUser = viewer
	}

	// I/O independent requests are launched together. This is especially important
	// with Supabase because network RTT is much larger than the Go processing time.
	var posts, pinned, hot []model.Post
	var blockedIDs []int
	var allAds []model.Advertisement
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() { defer wg.Done(); fn() }()
	}
	if blockedCategory == nil && (catSlug != "" || q != "") {
		run(func() { posts, _ = h.content.ListPostsLimited(q, catSlug, 60) })
	}
	if catSlug == "" {
		run(func() { pinned, _ = h.content.ListPinnedPosts("", 8) })
		if q == "" {
			// Trang chủ hiển thị toàn bộ bài viết, mới nhất ở trên; không giới hạn trong hôm nay.
			run(func() { posts, _ = h.content.ListPosts("", "") })
		}
		run(func() { hot, _ = h.content.ListTrendingPosts(40) })
	}
	if viewer != nil {
		run(func() { blockedIDs, _ = h.content.ListBlockedUserIDs(viewer.ID) })
	}
	run(func() { allAds, _ = h.content.ListAds("", true) })
	wg.Wait()

	blockedAuthors := map[int]bool{}
	for _, id := range blockedIDs {
		blockedAuthors[id] = true
	}
	authors := h.loadAuthors(posts, pinned, hot)
	getAuthor := func(id int) *model.User { return authors[id] }

	filterPosts := func(items []model.Post) []model.Post {
		out := make([]model.Post, 0, len(items))
		for _, post := range items {
			if blockedAuthors[post.AuthorID] {
				continue
			}
			if !h.postVisibleTo(post, allCats, viewer, getAuthor(post.AuthorID)) {
				continue
			}
			if len(activeFilterValues) > 0 {
				var pm model.PostMeta
				_ = json.Unmarshal([]byte(post.MetaJSON), &pm)
				matched := true
				for key, want := range activeFilterValues {
					if want != "" && !strings.EqualFold(metaFilterValue(pm, key), want) {
						matched = false
						break
					}
				}
				if !matched {
					continue
				}
			}
			out = append(out, post)
		}
		return out
	}
	withoutDocuments := func(items []model.Post) []model.Post {
		if catSlug != "" {
			return items
		}
		out := make([]model.Post, 0, len(items))
		for _, post := range items {
			if categoryStyle[post.CategoryID] != "document" {
				out = append(out, post)
			}
		}
		return out
	}
	trim := func(items []model.Post, n int) []model.Post {
		if n > 0 && len(items) > n {
			return items[:n]
		}
		return items
	}

	filteredPosts := filterPosts(posts)
	data.Posts = filteredPosts
	pinnedShown := []model.Post{}
	todayShown := []model.Post{}
	hotShown := []model.Post{}
	if catSlug == "" {
		pinnedShown = trim(withoutDocuments(filterPosts(pinned)), 4)
		hotShown = trim(withoutDocuments(filterPosts(hot)), 5)
	}
	if catSlug != "" || q != "" {
		todayShown = trim(filteredPosts, 60)
	} else {
		// Không cắt theo ngày/8 bài: toàn bộ bài thường được giữ, theo thứ tự mới -> cũ
		// do repository đã ORDER BY created_at DESC, id DESC.
		todayShown = withoutDocuments(filteredPosts)
	}

	// Stats and saved-state are fetched once for every card on the page instead of
	// once per column. This removes four redundant Supabase round-trips.
	ids := postIDs(pinnedShown, todayShown, hotShown)
	statsByID := map[int]model.PostStats{}
	savedByID := map[int]bool{}
	var cardWG sync.WaitGroup
	cardWG.Add(1)
	go func() {
		defer cardWG.Done()
		statsByID, _ = h.content.GetPostStatsBatch(ids)
	}()
	if currentUserID > 0 {
		cardWG.Add(1)
		go func() {
			defer cardWG.Done()
			savedByID, _ = h.content.ListSavedPostIDs(currentUserID, ids)
		}()
	}
	cardWG.Wait()
	makeCards := func(items []model.Post) []model.PostCard {
		cards := make([]model.PostCard, 0, len(items))
		for _, post := range items {
			var meta model.PostMeta
			_ = json.Unmarshal([]byte(post.MetaJSON), &meta)
			cards = append(cards, model.PostCard{Post: post, Meta: meta, Stats: statsByID[post.ID], Saved: savedByID[post.ID], Author: getAuthor(post.AuthorID), CardStyle: categoryStyle[post.CategoryID]})
		}
		return cards
	}
	data.Pinned = makeCards(pinnedShown)
	data.TodayPosts = makeCards(todayShown)
	data.HotPosts = makeCards(hotShown)
	data.Trending = data.HotPosts

	for _, ad := range allAds {
		switch ad.Position {
		case "home-right":
			data.HomeAds = append(data.HomeAds, ad)
		case "feed-sponsored":
			data.FeedAds = append(data.FeedAds, ad)
		}
	}
	h.render(w, "home.html", data)
}

func (h *Handler) PostDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/bai-viet/"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	p, err := h.content.GetPost(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	allCats, _ := h.content.ListCategories()
	var viewer *model.User
	if u, ok := h.sessions.CurrentUser(r); ok {
		viewer = &u
	}
	var author *model.User
	if p.AuthorID > 0 {
		if users, err := h.auth.GetUsersByIDs([]int{p.AuthorID}); err == nil {
			if u, ok := users[p.AuthorID]; ok {
				copy := u
				author = &copy
			}
		}
	}
	if !h.postVisibleTo(p, allCats, viewer, author) {
		if cat, found := findCategory(allCats, p.CategoryID); found {
			message := ""
			if audienceScopes(cat.AudienceScope)["same_school"] && viewer != nil && strings.TrimSpace(viewer.School) != "" && author != nil && strings.TrimSpace(author.School) != "" && !reference.SameSchool(viewer.School, author.School) {
				message = "Nội dung này chỉ dành cho sinh viên cùng trường với người đăng. Hãy kiểm tra lại Trường / Đại học trong hồ sơ của bạn."
			}
			h.redirectAudienceAccess(w, r, cat, viewer, message)
			return
		}
		http.Error(w, "Bạn không có quyền xem bài viết này", http.StatusForbidden)
		return
	}

	var meta model.PostMeta
	_ = json.Unmarshal([]byte(p.MetaJSON), &meta)
	data := model.PageData{Title: p.Title, Post: &p, PostMeta: meta, Categories: allCats, Category: p.CategorySlug, Message: r.URL.Query().Get("msg"), Author: author}
	if cat, found := findCategory(allCats, p.CategoryID); found {
		data.PostCategory = &cat
		data.PostCustomFields = customFieldDisplayList(cat, meta)
	}
	if viewer != nil {
		data.IsLoggedIn = true
		data.UserName = viewer.Name
		data.CurrentUser = viewer
		data.IsOwnPost = viewer.IsAdmin || (p.AuthorID > 0 && p.AuthorID == viewer.ID)
	}

	isDocumentDetail := data.PostCategory != nil && data.PostCategory.PostCardStyle == "document"
	var mainStats model.PostStats
	var suggested []model.Post
	var comments []model.Comment
	var allAds []model.Advertisement
	var blockedAuthor bool
	var userVote int
	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() { defer wg.Done(); fn() }()
	}
	run(func() { mainStats, _ = h.content.GetPostStats(p.ID) })
	run(func() { suggested, _ = h.content.ListPostsLimited("", p.CategorySlug, 12) })
	run(func() { comments, _ = h.content.ListComments(p.ID) })
	run(func() { allAds, _ = h.content.ListAds("", true) })
	if viewer != nil {
		if p.AuthorID > 0 && p.AuthorID != viewer.ID {
			run(func() { blockedAuthor, _ = h.content.IsUserBlocked(viewer.ID, p.AuthorID) })
		}
		run(func() { userVote, _ = h.content.GetUserPostVote(p.ID, viewer.ID) })
	}
	wg.Wait()
	data.PostStats = mainStats
	data.BlockedAuthor = blockedAuthor
	data.UserPostVote = userVote
	data.Comments = service.BuildCommentTree(comments)

	if len(suggested) < 8 || isDocumentDetail {
		all, _ := h.content.ListPostsLimited("", "", 20)
		suggested = append(suggested, all...)
	}
	suggestedAuthors := h.loadAuthors(suggested)
	seen := map[int]bool{p.ID: true}
	selectedPosts := make([]model.Post, 0, 5)
	selectedMeta := make(map[int]model.PostMeta, 5)
	for _, sp := range suggested {
		if seen[sp.ID] {
			continue
		}
		seen[sp.ID] = true
		var sm model.PostMeta
		_ = json.Unmarshal([]byte(sp.MetaJSON), &sm)
		spCategory, spCategoryFound := findCategory(allCats, sp.CategoryID)
		isSuggestedDocument := spCategoryFound && spCategory.PostCardStyle == "document" && strings.TrimSpace(sm.DocumentFile) != ""
		if isDocumentDetail {
			if !isSuggestedDocument {
				continue
			}
		} else if isSuggestedDocument {
			continue
		}
		if !h.postVisibleTo(sp, allCats, viewer, suggestedAuthors[sp.AuthorID]) {
			continue
		}
		selectedPosts = append(selectedPosts, sp)
		selectedMeta[sp.ID] = sm
		if len(selectedPosts) >= 5 {
			break
		}
	}
	statsByID, _ := h.content.GetPostStatsBatch(postIDs(selectedPosts))
	cards := make([]model.PostCard, 0, len(selectedPosts))
	for _, sp := range selectedPosts {
		cards = append(cards, model.PostCard{Post: sp, Meta: selectedMeta[sp.ID], Author: suggestedAuthors[sp.AuthorID], Stats: statsByID[sp.ID]})
	}
	data.Suggested = cards
	for _, ad := range allAds {
		switch ad.Position {
		case "post-left":
			if data.LeftAd == nil {
				copy := ad
				data.LeftAd = &copy
			}
		case "post-right":
			if data.RightAd == nil {
				copy := ad
				data.RightAd = &copy
			}
		}
	}
	h.render(w, "post.html", data)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")
		u, err := h.auth.Login(login, password)
		if err == nil {
			h.sessions.LoginUser(w, u)
			http.Redirect(w, r, safeNext(r.FormValue("next"), "/"), 303)
			return
		}
		h.render(w, "login.html", h.withHeaderSession(r, model.PageData{Title: "Đăng nhập", Message: err.Error(), AuthMode: "login", GoogleClientID: h.auth.GoogleClientID(), Query: r.FormValue("next")}))
		return
	}
	h.render(w, "login.html", h.withHeaderSession(r, model.PageData{Title: "Đăng nhập", AuthMode: "login", GoogleClientID: h.auth.GoogleClientID(), Query: r.URL.Query().Get("next")}))
}
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if r.FormValue("password") != r.FormValue("confirm_password") {
			h.render(w, "login.html", h.withHeaderSession(r, model.PageData{Title: "Đăng ký", Message: "Mật khẩu xác nhận không khớp", AuthMode: "register", GoogleClientID: h.auth.GoogleClientID(), Query: r.FormValue("next")}))
			return
		}
		u, err := h.auth.Register(r.FormValue("phone"), r.FormValue("password"))
		if err == nil {
			h.sessions.LoginUser(w, u)
			http.Redirect(w, r, safeNext(r.FormValue("next"), "/"), 303)
			return
		}
		h.render(w, "login.html", h.withHeaderSession(r, model.PageData{Title: "Đăng ký", Message: err.Error(), AuthMode: "register", GoogleClientID: h.auth.GoogleClientID(), Query: r.FormValue("next")}))
		return
	}
	h.render(w, "login.html", h.withHeaderSession(r, model.PageData{Title: "Đăng ký", AuthMode: "register", GoogleClientID: h.auth.GoogleClientID(), Query: r.URL.Query().Get("next")}))
}
func (h *Handler) PhoneAccountCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	exists, _, err := h.auth.PhoneAccountStatus(r.URL.Query().Get("phone"))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "exists": false, "message": err.Error()})
		return
	}
	payload := map[string]any{"ok": true, "exists": exists}
	if exists {
		payload["message"] = "Số điện thoại đã có tài khoản"
	} else {
		payload["message"] = "Số điện thoại có thể đăng ký"
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var body struct {
		Credential string `json:"credential"`
		Next       string `json:"next"`
	}
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		http.Error(w, "Dữ liệu không hợp lệ", 400)
		return
	}
	u, err := h.auth.LoginWithGoogle(body.Credential)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
		return
	}
	h.sessions.LoginUser(w, u)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "redirect": safeNext(body.Next, "/")})
}
func (h *Handler) ToggleSavePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, ok := h.sessions.CurrentUser(r)
	isAJAX := r.Header.Get("X-Requested-With") == "XMLHttpRequest" || strings.Contains(r.Header.Get("Accept"), "application/json")
	if !ok {
		redirect := "/login?next=" + safeNext(r.FormValue("next"), "/")
		if isAJAX {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "redirect": redirect})
			return
		}
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}
	if rejectRestricted(w, r, u) {
		return
	}
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	if !h.userCanAccessPost(postID, &u) {
		http.Error(w, "Bạn không có quyền truy cập bài viết này", http.StatusForbidden)
		return
	}
	saved, err := h.content.ToggleSavedPost(postID, u.ID)
	if err != nil {
		if isAJAX {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
			return
		}
		http.Redirect(w, r, safeNext(r.FormValue("next"), "/"), http.StatusSeeOther)
		return
	}
	if isAJAX {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "saved": saved})
		return
	}
	http.Redirect(w, r, safeNext(r.FormValue("next"), "/bai-da-luu"), http.StatusSeeOther)
}

func (h *Handler) VotePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	isAJAX := r.Header.Get("X-Requested-With") == "XMLHttpRequest" || strings.Contains(r.Header.Get("Accept"), "application/json")
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		next := safeNext(r.FormValue("next"), "/")
		loginURL := "/login?next=" + next
		if isAJAX {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Bạn cần đăng nhập để bình chọn.", "redirect": loginURL})
			return
		}
		http.Redirect(w, r, loginURL, 303)
		return
	}
	if rejectRestricted(w, r, u) {
		return
	}
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	value, _ := strconv.Atoi(r.FormValue("value"))
	if !h.userCanAccessPost(postID, &u) {
		http.Error(w, "Bạn không có quyền truy cập bài viết này", http.StatusForbidden)
		return
	}
	if err := h.content.VotePost(postID, u.ID, value); err != nil {
		if isAJAX {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
			return
		}
		http.Redirect(w, r, safeNext(r.FormValue("next"), "/"), 303)
		return
	}
	if isAJAX {
		stats, err := h.content.GetPostStats(postID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Không thể cập nhật điểm bình chọn."})
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":        true,
			"post_id":   postID,
			"user_vote": value,
			"score":     stats.Score,
			"upvotes":   stats.Upvotes,
			"downvotes": stats.Downvotes,
			"comments":  stats.Comments,
		})
		return
	}
	http.Redirect(w, r, safeNext(r.FormValue("next"), "/"), 303)
}
func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		next := safeNext(r.FormValue("next"), "/")
		http.Redirect(w, r, "/login?next="+next, 303)
		return
	}
	if rejectRestricted(w, r, u) {
		return
	}
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	if !h.userCanAccessPost(postID, &u) {
		http.Error(w, "Bạn không có quyền truy cập bài viết này", http.StatusForbidden)
		return
	}
	parentID, _ := strconv.Atoi(r.FormValue("parent_id"))
	content := strings.TrimSpace(r.FormValue("content"))

	if parentID > 0 {
		comments, _ := h.content.ListComments(postID)
		byID := map[int]model.Comment{}
		for _, item := range comments {
			byID[item.ID] = item
		}
		parent, found := byID[parentID]
		if !found || parent.PostID != postID {
			http.Redirect(w, r, safeNext(r.FormValue("next"), "/")+"#comments", 303)
			return
		}
		depth := 0
		cursor := parent
		for cursor.ParentID != 0 {
			depth++
			ancestor, ok := byID[cursor.ParentID]
			if !ok {
				break
			}
			cursor = ancestor
		}
		if depth >= 3 {
			http.Redirect(w, r, safeNext(r.FormValue("next"), "/")+"#comments", 303)
			return
		}
		mention := "@" + strings.TrimSpace(parent.UserName)
		if mention != "@" && !strings.HasPrefix(strings.TrimSpace(content), mention) {
			content = mention + " " + content
		}
	}

	if err := h.content.AddComment(model.Comment{PostID: postID, ParentID: parentID, UserID: u.ID, Content: content}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, safeNext(r.FormValue("next"), "/")+"#comments", 303)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sessions.Logout(w, r)
	http.Redirect(w, r, "/", 303)
}

func (h *Handler) profileCards(user model.User, viewer *model.User, cats []model.Category) []model.PostCard {
	posts, _ := h.content.ListPostsByAuthor(user.ID)
	visible := make([]model.Post, 0, len(posts))
	metaByID := make(map[int]model.PostMeta, len(posts))
	author := user
	for _, post := range posts {
		if !h.postVisibleTo(post, cats, viewer, &author) {
			continue
		}
		var meta model.PostMeta
		_ = json.Unmarshal([]byte(post.MetaJSON), &meta)
		visible = append(visible, post)
		metaByID[post.ID] = meta
	}
	statsByID, _ := h.content.GetPostStatsBatch(postIDs(visible))
	cards := make([]model.PostCard, 0, len(visible))
	for _, post := range visible {
		cards = append(cards, model.PostCard{Post: post, Meta: metaByID[post.ID], Stats: statsByID[post.ID], Author: &author})
	}
	return cards
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login?next=/ho-so", 303)
		return
	}
	cats, _ := h.content.ListCategories()
	message := r.URL.Query().Get("msg")
	next := safeNext(r.URL.Query().Get("next"), "")
	if r.Method == http.MethodPost {
		next = safeNext(r.FormValue("next"), "")
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			message = "Ảnh tải lên quá lớn hoặc dữ liệu không hợp lệ"
		} else {
			avatarURL, err := saveUpload(r, "avatar")
			if err != nil {
				message = err.Error()
			} else {
				if avatarURL == "" {
					avatarURL = u.AvatarURL
				}
				_, err := h.auth.UpdateProfile(u.ID, r.FormValue("name"), u.Username, avatarURL, r.FormValue("profile_role"), r.FormValue("school"), r.FormValue("major"), r.FormValue("student_id"), r.FormValue("phone"), r.FormValue("employer_company"), r.FormValue("employer_tax_code"), r.FormValue("employer_representative"), r.FormValue("employer_website"), r.FormValue("landlord_name"), r.FormValue("landlord_address"), r.FormValue("landlord_phone"), r.FormValue("landlord_legal_info"))
				if err != nil {
					message = err.Error()
				} else {
					if next != "" {
						http.Redirect(w, r, next, 303)
						return
					}
					http.Redirect(w, r, "/ho-so?msg="+urlMessage("Đã cập nhật hồ sơ"), 303)
					return
				}
			}
		}
	}
	if latest, err := h.auth.GetUserByID(u.ID); err == nil {
		u = latest
	}
	profileUser := u
	visibleCats := cats
	var verificationReq *model.VerificationRequest
	var phonePending bool
	var profilePosts []model.PostCard
	var profileWG sync.WaitGroup
	profileWG.Add(3)
	go func() {
		defer profileWG.Done()
		if vr, err := h.content.GetLatestVerificationRequest(u.ID); err == nil {
			verificationReq = &vr
		}
	}()
	go func() {
		defer profileWG.Done()
		phonePending, _ = h.content.HasPendingVerificationRequest(u.ID, "phone")
	}()
	go func() {
		defer profileWG.Done()
		profilePosts = h.profileCards(profileUser, &u, cats)
	}()
	profileWG.Wait()
	h.render(w, "profile.html", model.PageData{
		Title:                    "Hồ sơ",
		Categories:               visibleCats,
		IsLoggedIn:               true,
		UserName:                 u.Name,
		CurrentUser:              &u,
		ProfileUser:              &profileUser,
		ProfilePosts:             profilePosts,
		IsOwnProfile:             true,
		Message:                  message,
		ReturnTo:                 next,
		VerificationRequest:      verificationReq,
		EducationEmailOK:         isEducationEmail(u.Email),
		PhoneVerificationPending: phonePending,
	})
}

func (h *Handler) PublicProfile(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/nguoi-dung/"))
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	profileUser, err := h.auth.GetUserByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	allCats, _ := h.content.ListCategories()
	var viewer *model.User
	if current, ok := h.sessions.CurrentUser(r); ok {
		viewer = &current
	}
	data := model.PageData{
		Title:        profileUser.Name,
		Categories:   allCats,
		ProfileUser:  &profileUser,
		ProfilePosts: h.profileCards(profileUser, viewer, allCats),
		UserName:     profileUser.Name,
	}
	if viewer != nil {
		data.IsLoggedIn = true
		data.CurrentUser = viewer
		data.UserName = viewer.Name
		data.IsOwnProfile = viewer.ID == profileUser.ID
	}
	h.render(w, "profile.html", data)
}

func (h *Handler) RequestPhoneVerification(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login?next=/ho-so", 303)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	latest, err := h.auth.GetUserByID(u.ID)
	if err == nil {
		u = latest
	}
	if submittedPhone := strings.TrimSpace(r.FormValue("phone")); submittedPhone != "" && submittedPhone != strings.TrimSpace(u.Phone) {
		updated, updateErr := h.auth.UpdateProfile(u.ID, u.Name, u.Username, u.AvatarURL, u.ProfileRole, u.School, u.Major, u.StudentID, submittedPhone, u.EmployerCompany, u.EmployerTaxCode, u.EmployerRepresentative, u.EmployerWebsite, u.LandlordName, u.LandlordAddress, u.LandlordPhone, u.LandlordLegalInfo)
		if updateErr != nil {
			http.Redirect(w, r, "/ho-so?msg="+urlMessage(updateErr.Error()), 303)
			return
		}
		u = updated
	}
	if u.PhoneVerified {
		http.Redirect(w, r, "/ho-so?msg="+urlMessage("Số điện thoại đã được xác thực"), 303)
		return
	}
	if strings.TrimSpace(u.Phone) == "" {
		http.Redirect(w, r, "/ho-so?msg="+urlMessage("Vui lòng lưu số điện thoại trong hồ sơ trước"), 303)
		return
	}
	if pending, err := h.content.HasPendingVerificationRequest(u.ID, "phone"); err == nil && pending {
		http.Redirect(w, r, "/ho-so?msg="+urlMessage("Yêu cầu xác thực số điện thoại đang chờ duyệt"), 303)
		return
	}
	info := "Số điện thoại: " + strings.TrimSpace(u.Phone)
	if err := h.content.CreateVerificationRequest(model.VerificationRequest{UserID: u.ID, Type: "phone", Info: info}); err != nil {
		http.Redirect(w, r, "/ho-so?msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/ho-so?msg="+urlMessage("Đã gửi yêu cầu xác thực số điện thoại"), 303)
}

func (h *Handler) RequestVerification(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login?next=/ho-so", 303)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if latest, err := h.auth.GetUserByID(u.ID); err == nil {
		u = latest
	}
	if latest, err := h.content.GetLatestVerificationRequest(u.ID); err == nil && latest.Status == "pending" && latest.Type != "phone" {
		http.Redirect(w, r, "/ho-so?msg="+urlMessage("Yêu cầu xác thực vai trò đang chờ quản trị viên duyệt"), 303)
		return
	}
	verificationType := strings.TrimSpace(r.FormValue("type"))
	if verificationType == "student" && !isEducationEmail(u.Email) && !(u.Provider == "phone" && strings.TrimSpace(u.Phone) != "") {
		http.Redirect(w, r, "/ho-so?msg="+urlMessage("Xác thực Sinh viên yêu cầu email giáo dục hoặc tài khoản đăng ký bằng số điện thoại"), 303)
		return
	}
	verificationInfo, infoErr := buildVerificationInfoForUser(r, verificationType, u)
	if infoErr != nil {
		http.Redirect(w, r, "/ho-so?msg="+urlMessage(infoErr.Error()), 303)
		return
	}
	err := h.content.CreateVerificationRequest(model.VerificationRequest{UserID: u.ID, Type: verificationType, Info: verificationInfo})
	if err != nil {
		http.Redirect(w, r, "/ho-so?msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/ho-so?msg="+urlMessage("Đã gửi yêu cầu xác thực"), 303)
}

func buildVerificationInfoForUser(r *http.Request, verificationType string, u model.User) (string, error) {
	join := func(rows ...string) string { return strings.Join(rows, "\n") }

	switch verificationType {
	case "student":
		phoneIdentity := u.Provider == "phone" && strings.TrimSpace(u.Phone) != ""
		if !isEducationEmail(u.Email) && !phoneIdentity {
			return "", errors.New("xác thực Sinh viên yêu cầu email giáo dục hoặc tài khoản số điện thoại")
		}
		if strings.TrimSpace(u.School) == "" || strings.TrimSpace(u.Major) == "" || strings.TrimSpace(u.StudentID) == "" {
			return "", errors.New("vui lòng cập nhật đầy đủ thông tin Sinh viên trong Thông tin cá nhân trước")
		}
		identity := "Email giáo dục: " + u.Email
		if phoneIdentity {
			identity = "Tài khoản số điện thoại: " + u.Phone
		}
		return join(identity, "Trường / Đại học: "+u.School, "Ngành học: "+u.Major, "Mã số sinh viên: "+u.StudentID, "Số điện thoại hồ sơ: "+u.Phone), nil
	case "employer":
		if strings.TrimSpace(u.EmployerCompany) == "" || strings.TrimSpace(u.EmployerTaxCode) == "" || strings.TrimSpace(u.EmployerRepresentative) == "" || strings.TrimSpace(u.EmployerWebsite) == "" {
			return "", errors.New("vui lòng cập nhật đầy đủ Thông tin Nhà tuyển dụng trong Thông tin cá nhân trước")
		}
		return join("Công ty / đơn vị: "+u.EmployerCompany, "Mã số thuế / mã doanh nghiệp: "+u.EmployerTaxCode, "Chức vụ / người đại diện: "+u.EmployerRepresentative, "Website / Fanpage: "+u.EmployerWebsite, "Số điện thoại hồ sơ: "+u.Phone), nil
	case "landlord":
		if strings.TrimSpace(u.LandlordName) == "" || strings.TrimSpace(u.LandlordAddress) == "" || strings.TrimSpace(u.LandlordPhone) == "" || strings.TrimSpace(u.LandlordLegalInfo) == "" {
			return "", errors.New("vui lòng cập nhật đầy đủ Thông tin Chủ trọ trong Thông tin cá nhân trước")
		}
		return join("Chủ trọ / cơ sở: "+u.LandlordName, "Địa chỉ khu trọ: "+u.LandlordAddress, "Số điện thoại liên hệ: "+u.LandlordPhone, "Thông tin giấy tờ xác minh: "+u.LandlordLegalInfo, "Số điện thoại hồ sơ: "+u.Phone), nil
	default:
		return "", errors.New("loại xác thực không hợp lệ")
	}
}

func (h *Handler) AdminVerificationAction(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin?tab=verification&msg="+urlMessage("Dữ liệu xác thực không hợp lệ"), 303)
		return
	}
	id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("id")))
	if err != nil || id <= 0 {
		msg := "Không xác định được yêu cầu xác thực"
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": msg})
			return
		}
		http.Redirect(w, r, "/admin?tab=verification&msg="+urlMessage(msg), 303)
		return
	}
	action := strings.TrimSpace(r.FormValue("action"))
	status := ""
	switch action {
	case "approve":
		status = "approved"
	case "reject":
		status = "rejected"
	default:
		msg := "Thao tác xác thực không hợp lệ"
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": msg})
			return
		}
		http.Redirect(w, r, "/admin?tab=verification&msg="+urlMessage(msg), 303)
		return
	}
	if err := h.content.ResolveVerificationRequest(id, status); err != nil {
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		http.Redirect(w, r, "/admin?tab=verification&msg="+urlMessage(err.Error()), 303)
		return
	}
	if wantsJSON(r) {
		writeAdminJSON(w, 200, map[string]any{"ok": true, "status": status})
		return
	}
	http.Redirect(w, r, "/admin?tab=verification&msg="+urlMessage("Đã cập nhật yêu cầu xác thực"), 303)
}

func isEducationEmail(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := parts[1]
	return strings.HasSuffix(domain, ".edu.vn") || strings.HasSuffix(domain, ".edu") || strings.Contains(domain, ".edu.")
}

func safeNext(next, fallback string) string {
	next = strings.TrimSpace(next)
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return fallback
	}
	return next
}

func findCategory(cats []model.Category, id int) (model.Category, bool) {
	for _, c := range cats {
		if c.ID == id {
			return c, true
		}
	}
	return model.Category{}, false
}
func enabledConfig(raw string) []model.FieldConfig {
	var out []model.FieldConfig
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}
func required(cfg []model.FieldConfig, key string) bool {
	for _, f := range cfg {
		if f.Key == key && f.Enabled {
			return f.Required
		}
	}
	return false
}
func enabled(cfg []model.FieldConfig, key string) bool {
	for _, f := range cfg {
		if f.Key == key {
			return f.Enabled
		}
	}
	return false
}

func isCustomFieldConfig(f model.FieldConfig) bool {
	return strings.HasPrefix(strings.TrimSpace(f.Key), "custom_")
}

func collectCustomFields(r *http.Request, cfg []model.FieldConfig) map[string]string {
	out := map[string]string{}
	for _, f := range cfg {
		if !f.Enabled || !isCustomFieldConfig(f) {
			continue
		}
		value := strings.TrimSpace(r.FormValue(f.Key))
		if value != "" {
			out[f.Key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func validateCustomFields(meta model.PostMeta, cfg []model.FieldConfig) error {
	for _, f := range cfg {
		if !f.Enabled || !isCustomFieldConfig(f) {
			continue
		}
		value := strings.TrimSpace(meta.CustomFields[f.Key])
		if f.Required && value == "" {
			return fmt.Errorf("Vui lòng nhập trường bắt buộc: %s", f.Label)
		}
		if len([]rune(value)) > 4000 {
			return fmt.Errorf("%s quá dài", f.Label)
		}
		switch strings.ToLower(strings.TrimSpace(f.Type)) {
		case "email":
			if value != "" {
				addr, err := mail.ParseAddress(value)
				if err != nil || !strings.EqualFold(addr.Address, value) {
					return fmt.Errorf("%s không phải email hợp lệ", f.Label)
				}
			}
		case "url":
			if value != "" {
				u, err := url.ParseRequestURI(value)
				if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
					return fmt.Errorf("%s không phải liên kết hợp lệ", f.Label)
				}
			}
		}
	}
	return nil
}

func customFieldValuesJSON(meta model.PostMeta) string {
	if meta.CustomFields == nil {
		return "{}"
	}
	b, _ := json.Marshal(meta.CustomFields)
	return string(b)
}

func customFieldDisplayList(cat model.Category, meta model.PostMeta) []model.PostCustomField {
	if len(meta.CustomFields) == 0 {
		return nil
	}
	out := []model.PostCustomField{}
	for _, f := range enabledConfig(cat.FormConfig) {
		if !f.Enabled || !isCustomFieldConfig(f) {
			continue
		}
		if value := strings.TrimSpace(meta.CustomFields[f.Key]); value != "" {
			out = append(out, model.PostCustomField{Key: f.Key, Label: f.Label, Value: value})
		}
	}
	return out
}

func uploadDir() string {
	if v := strings.TrimSpace(os.Getenv("UPLOAD_DIR")); v != "" {
		return v
	}
	return "uploads"
}

func supabaseStorageConfig() (baseURL, serviceKey, bucket string, ok bool) {
	baseURL = strings.TrimRight(strings.TrimSpace(os.Getenv("SUPABASE_URL")), "/")
	serviceKey = strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
	bucket = strings.TrimSpace(os.Getenv("SUPABASE_STORAGE_BUCKET"))
	if bucket == "" {
		bucket = "uploads"
	}
	ok = baseURL != "" && serviceKey != ""
	return
}

func storeUpload(data []byte, name, contentType string) (string, error) {
	if baseURL, key, bucket, ok := supabaseStorageConfig(); ok {
		endpoint := baseURL + "/storage/v1/object/" + url.PathEscape(bucket) + "/" + url.PathEscape(name)
		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("apikey", key)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Cache-Control", "3600")
		client := &http.Client{Timeout: 45 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("không thể tải tệp lên Supabase Storage: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			return "", fmt.Errorf("Supabase Storage từ chối tải tệp (%d): %s", resp.StatusCode, strings.TrimSpace(string(msg)))
		}
		return baseURL + "/storage/v1/object/public/" + url.PathEscape(bucket) + "/" + url.PathEscape(name), nil
	}

	dir := uploadDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

func deleteStoredUpload(fileURL string) error {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return nil
	}
	if strings.HasPrefix(fileURL, "/uploads/") {
		rel := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(fileURL, "/uploads/")))
		if rel == "." || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil
		}
		err := os.Remove(filepath.Join(uploadDir(), rel))
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	baseURL, key, bucket, ok := supabaseStorageConfig()
	if !ok {
		return nil
	}
	parsed, err := url.Parse(fileURL)
	if err != nil || !strings.HasPrefix(fileURL, baseURL+"/") {
		return nil
	}
	marker := "/storage/v1/object/public/" + url.PathEscape(bucket) + "/"
	if !strings.HasPrefix(parsed.EscapedPath(), marker) && !strings.HasPrefix(parsed.Path, marker) {
		return nil
	}
	pathValue := parsed.EscapedPath()
	if !strings.HasPrefix(pathValue, marker) {
		pathValue = parsed.Path
	}
	name := strings.TrimPrefix(pathValue, marker)
	name, _ = url.PathUnescape(name)
	name = strings.TrimSpace(name)
	if name == "" || strings.Contains(name, "../") || strings.HasPrefix(name, "/") {
		return nil
	}
	endpoint := baseURL + "/storage/v1/object/" + url.PathEscape(bucket) + "/" + url.PathEscape(name)
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("apikey", key)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("không thể xóa ảnh khỏi Storage (%d): %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}

func readLimitedUpload(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("tệp vượt quá dung lượng cho phép")
	}
	return data, nil
}

func materializeDocument(fileURL string) (string, error) {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return "", errors.New("tài liệu trống")
	}
	if strings.HasPrefix(fileURL, "/uploads/") {
		rel := strings.TrimPrefix(fileURL, "/uploads/")
		return filepath.Join(uploadDir(), filepath.FromSlash(rel)), nil
	}
	if !strings.HasPrefix(fileURL, "http://") && !strings.HasPrefix(fileURL, "https://") {
		return fileURL, nil
	}
	parsed, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(parsed.Path))
	cacheDir := filepath.Join(uploadDir(), "remote-cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}
	hash := sha1.Sum([]byte(fileURL))
	path := filepath.Join(cacheDir, hex.EncodeToString(hash[:])+ext)
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return path, nil
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("không thể tải tài liệu từ Storage: HTTP %d", resp.StatusCode)
	}
	data, err := readLimitedUpload(resp.Body, 20<<20)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func saveUpload(r *http.Request, field string) (string, error) {
	file, header, err := r.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		return "", errors.New("ảnh chỉ hỗ trợ JPG, PNG hoặc WEBP")
	}
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	name := hex.EncodeToString(buf) + ext
	data, err := readLimitedUpload(file, 8<<20)
	if err != nil {
		return "", err
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	return storeUpload(data, name, contentType)
}
func normalizeAllowedDocumentFormats(raw string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(service.NormalizeDocumentFormats(raw), ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part != "" {
			out["."+part] = true
		}
	}
	return out
}
func documentFormatAllowed(fileURL, allowedRaw string) bool {
	if strings.TrimSpace(fileURL) == "" {
		return true
	}
	ext := strings.ToLower(filepath.Ext(fileURL))
	return normalizeAllowedDocumentFormats(allowedRaw)[ext]
}
func saveDocumentUpload(r *http.Request, field, allowedRaw string) (string, error) {
	file, header, err := r.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := normalizeAllowedDocumentFormats(allowedRaw)
	if !allowed[ext] {
		return "", errors.New("định dạng tài liệu này chưa được cho phép trong chuyên mục")
	}
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	name := hex.EncodeToString(buf) + ext
	data, err := readLimitedUpload(file, 16<<20)
	if err != nil {
		return "", err
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return storeUpload(data, name, contentType)
}

func previewCacheKey(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	hash := sha1.Sum([]byte(path + "|" + strconv.FormatInt(info.ModTime().UnixNano(), 10) + "|" + strconv.FormatInt(info.Size(), 10)))
	return hex.EncodeToString(hash[:]), nil
}

func officeConverter() string {
	for _, name := range []string{"soffice", "libreoffice"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// ensurePreviewPDFFromDocument creates a cached PDF only for preview purposes.
// The original uploaded document is never modified.
func ensurePreviewPDFFromDocument(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".pdf" {
		if strings.HasPrefix(filepath.ToSlash(path), "uploads/") {
			return "/" + filepath.ToSlash(path)
		}
		return ""
	}
	if ext == ".zip" {
		return ""
	}
	converter := officeConverter()
	if converter == "" {
		return ""
	}
	key, err := previewCacheKey(path)
	if err != nil {
		return ""
	}
	previewDir := filepath.Join(uploadDir(), "previews")
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return ""
	}
	cachedPDF := filepath.Join(previewDir, key+".pdf")
	if info, err := os.Stat(cachedPDF); err == nil && info.Size() > 0 {
		return "/uploads/previews/" + key + ".pdf"
	}

	tmpDir, err := os.MkdirTemp("", "dacs-preview-pdf-*")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tmpDir)
	profileDir, err := os.MkdirTemp("", "dacs-lo-profile-*")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(profileDir)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	profileURL := "file://" + filepath.ToSlash(profileDir)
	cmd := exec.CommandContext(ctx, converter,
		"--headless",
		"-env:UserInstallation="+profileURL,
		"--convert-to", "pdf",
		"--outdir", tmpDir,
		path,
	)
	if err := cmd.Run(); err != nil {
		return ""
	}

	var converted string
	entries, _ := os.ReadDir(tmpDir)
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".pdf") {
			converted = filepath.Join(tmpDir, entry.Name())
			break
		}
	}
	if converted == "" {
		return ""
	}
	data, err := os.ReadFile(converted)
	if err != nil || len(data) == 0 {
		return ""
	}
	if err := os.WriteFile(cachedPDF, data, 0644); err != nil {
		return ""
	}
	return "/uploads/previews/" + key + ".pdf"
}

func documentPreviewCoverHTML(name, ext, imageURL string) string {
	label := strings.ToUpper(strings.TrimPrefix(ext, "."))
	if label == "" {
		label = "FILE"
	}
	if imageURL != "" {
		return fmt.Sprintf(`<div class="page cover-page"><div class="cover-image"><img src="%s" alt="Ảnh minh họa tài liệu"><div class="cover-chip">%s</div><strong>%s</strong></div></div>`, template.HTMLEscapeString(imageURL), template.HTMLEscapeString(label), template.HTMLEscapeString(name))
	}
	return fmt.Sprintf(`<div class="page cover-page"><div class="file-cover"><div class="file-sheet"><span>%s</span><b>📄</b></div><strong>%s</strong><small>Chưa thể chuyển tài liệu sang PDF trên máy này.</small></div></div>`, template.HTMLEscapeString(label), template.HTMLEscapeString(name))
}

func (h *Handler) DocumentPreview(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/tai-lieu/xem-truoc/"))
	if id <= 0 {
		http.NotFound(w, r)
		return
	}
	post, err := h.content.GetPost(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	cats, _ := h.content.ListCategories()
	var viewer *model.User
	if u, ok := h.sessions.CurrentUser(r); ok {
		viewer = &u
	}
	var author *model.User
	if post.AuthorID > 0 {
		if a, e := h.auth.GetUserByID(post.AuthorID); e == nil {
			author = &a
		}
	}
	if !h.postVisibleTo(post, cats, viewer, author) {
		http.Error(w, "Bạn không có quyền xem tài liệu này", http.StatusForbidden)
		return
	}
	var meta model.PostMeta
	_ = json.Unmarshal([]byte(post.MetaJSON), &meta)
	if meta.DocumentFile == "" {
		http.NotFound(w, r)
		return
	}
	parsedURL, _ := url.Parse(meta.DocumentFile)
	name := filepath.Base(parsedURL.Path)
	ext := strings.ToLower(filepath.Ext(name))
	path, materializeErr := materializeDocument(meta.DocumentFile)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><style>*{box-sizing:border-box}html,body{margin:0;width:100%;height:100%;overflow:hidden;font:14px system-ui;background:#eef3f0;color:#243c34}.wrap{height:100%;display:flex;flex-direction:column}.native{display:block;flex:1;width:100%;min-height:0;border:0;background:#fff}.page{flex:1;min-height:0;padding:18px;overflow:auto}.cover-page{display:grid;place-items:center}.cover-image,.file-cover{width:min(620px,100%);padding:22px;border-radius:18px;background:#fff;box-shadow:0 12px 32px rgba(19,60,48,.12);text-align:center}.cover-image img{display:block;width:100%;max-height:520px;object-fit:contain;border-radius:13px;background:#f4f7f5}.cover-image strong,.file-cover strong{display:block;margin-top:14px;overflow-wrap:anywhere}.cover-chip{display:inline-flex;margin-top:12px;padding:5px 9px;border-radius:999px;background:#e9f4ef;color:#0a5b47;font-size:10px;font-weight:900}.file-sheet{width:170px;height:220px;margin:auto;border-radius:18px;background:linear-gradient(160deg,#fff,#f3f8f5);border:1px solid #dfeae5;display:grid;place-items:center;position:relative}.file-sheet b{font-size:54px}.file-sheet span{position:absolute;top:15px;right:15px;padding:5px 8px;border-radius:8px;background:#0a5b47;color:#fff;font-size:10px;font-weight:900}.file-cover small{display:block;margin-top:7px;color:#77857f}</style></head><body><div class="wrap">`))

	if ext == ".pdf" {
		fmt.Fprintf(w, `<object class="native" data="%s" type="application/pdf"><div class="page">Trình duyệt không hỗ trợ xem PDF trực tiếp.</div></object>`, template.HTMLEscapeString(meta.DocumentFile))
		w.Write([]byte(`</div></body></html>`))
		return
	}
	if materializeErr == nil {
		if pdfURL := ensurePreviewPDFFromDocument(path); pdfURL != "" {
			fmt.Fprintf(w, `<object class="native" data="%s" type="application/pdf"><div class="page">Không thể mở PDF xem trước.</div></object>`, template.HTMLEscapeString(pdfURL))
			w.Write([]byte(`</div></body></html>`))
			return
		}
	}
	w.Write([]byte(documentPreviewCoverHTML(name, ext, meta.ImageURL)))
	w.Write([]byte(`</div></body></html>`))
}

func (h *Handler) DocumentUploadPreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Phương thức không hợp lệ"})
		return
	}
	if _, ok := h.sessions.CurrentUser(r); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Vui lòng đăng nhập"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 17<<20)
	if err := r.ParseMultipartForm(17 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Tệp quá lớn hoặc dữ liệu không hợp lệ"})
		return
	}
	file, header, err := r.FormFile("document_file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Chưa chọn tài liệu"})
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".txt": true, ".zip": true}
	if !allowed[ext] {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Định dạng này không hỗ trợ xem trước tạm thời"})
		return
	}
	tmp, err := os.CreateTemp("", "dacs-doc-preview-*"+ext)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Không thể tạo bản xem trước"})
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err = io.Copy(tmp, io.LimitReader(file, 16<<20)); err != nil {
		tmp.Close()
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "Không thể đọc tài liệu"})
		return
	}
	_ = tmp.Close()

	result := map[string]any{"ok": true, "name": header.Filename, "ext": strings.TrimPrefix(ext, ".")}
	if pdfURL := ensurePreviewPDFFromDocument(tmpPath); pdfURL != "" {
		result["kind"] = "pdf"
		result["preview_url"] = pdfURL
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	result["kind"] = "cover"
	result["message"] = "Không thể chuyển tệp này sang PDF. Hiển thị bìa minh họa thay thế."
	_ = json.NewEncoder(w).Encode(result)
}

func renderSubmit(h *Handler, w http.ResponseWriter, cats []model.Category, u model.User, msg string, draft *model.Draft) {
	data := model.PageData{Title: "Đăng bài", Categories: cats, IsLoggedIn: true, UserName: u.Name, CurrentUser: &u, Message: msg, Draft: draft, CustomFieldValuesJSON: "{}"}
	if draft != nil {
		_ = json.Unmarshal([]byte(draft.MetaJSON), &data.DraftMeta)
		data.CustomFieldValuesJSON = customFieldValuesJSON(data.DraftMeta)
	}
	h.render(w, "submit.html", data)
}

func collectPositions(r *http.Request) []model.RecruitmentPosition {
	positions := []model.RecruitmentPosition{}
	names := r.Form["position_name[]"]
	for i, name := range names {
		if strings.TrimSpace(name) == "" {
			continue
		}
		get := func(key string) string {
			v := r.Form[key+"[]"]
			if i < len(v) {
				return strings.TrimSpace(v[i])
			}
			return ""
		}
		label := get("position_label")
		if label == "" {
			label = "Vị trí tuyển dụng số " + strconv.Itoa(i+1)
		}
		positions = append(positions, model.RecruitmentPosition{Label: label, Name: strings.TrimSpace(name), JobType: get("position_job_type"), Specialty: get("position_specialty"), Location: get("position_location"), Salary: get("position_salary"), Experience: get("position_experience"), Description: get("position_description"), Requirement: get("position_requirement"), Benefit: get("position_benefit")})
	}
	return positions
}

func (h *Handler) SubmitPost(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login?next=/dang-bai", 303)
		return
	}
	if rejectRestricted(w, r, u) {
		return
	}
	allCats, err := h.content.ListCategories()
	if err != nil {
		http.Error(w, "Không thể tải chuyên mục", 500)
		return
	}
	cats := filterAccessibleCategories(allCats, &u)
	if r.Method == http.MethodGet {
		var draft *model.Draft
		if id, _ := strconv.Atoi(r.URL.Query().Get("draft")); id > 0 {
			if d, err := h.content.GetDraft(id, u.ID); err == nil {
				draft = &d
			}
		}
		renderSubmit(h, w, cats, u, "", draft)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		renderSubmit(h, w, cats, u, "Dữ liệu tải lên quá lớn hoặc không hợp lệ", nil)
		return
	}

	cid, _ := strconv.Atoi(r.FormValue("category_id"))
	selectedCat, selectedFound := findCategory(cats, cid)
	allowedDocFormats := "pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip"
	if selectedFound {
		allowedDocFormats = selectedCat.DocumentFormats
	}
	draftID, _ := strconv.Atoi(r.FormValue("draft_id"))
	var oldDraft *model.Draft
	var oldMeta model.PostMeta
	if draftID > 0 {
		if d, err := h.content.GetDraft(draftID, u.ID); err == nil {
			oldDraft = &d
			_ = json.Unmarshal([]byte(d.MetaJSON), &oldMeta)
		}
	}
	imageURL, err := saveUpload(r, "image")
	if err != nil {
		renderSubmit(h, w, cats, u, err.Error(), oldDraft)
		return
	}
	logoURL, err := saveUpload(r, "company_logo")
	if err != nil {
		renderSubmit(h, w, cats, u, err.Error(), oldDraft)
		return
	}
	documentURL, docErr := saveDocumentUpload(r, "document_file", allowedDocFormats)
	if docErr != nil {
		renderSubmit(h, w, cats, u, docErr.Error(), oldDraft)
		return
	}
	if imageURL == "" {
		imageURL = oldMeta.ImageURL
	}
	if logoURL == "" {
		logoURL = oldMeta.CompanyLogo
	}
	if documentURL == "" {
		documentURL = oldMeta.DocumentFile
	}
	meta := model.PostMeta{ImageURL: imageURL, Deadline: strings.TrimSpace(r.FormValue("deadline")), CompanyLogo: logoURL, Website: strings.TrimSpace(r.FormValue("website")), Fanpage: strings.TrimSpace(r.FormValue("fanpage")), RecruitmentContent: strings.TrimSpace(r.FormValue("recruitment_content")), CVEmail: strings.TrimSpace(r.FormValue("cv_email")), Positions: collectPositions(r), ContactName: strings.TrimSpace(r.FormValue("contact_name")), ContactPhone: strings.TrimSpace(r.FormValue("contact_phone")), Organization: strings.TrimSpace(r.FormValue("organization")), Location: strings.TrimSpace(r.FormValue("location")), SalaryRange: strings.TrimSpace(r.FormValue("salary_range")), ApplicationLink: strings.TrimSpace(r.FormValue("application_link")), EventTime: strings.TrimSpace(r.FormValue("event_time")), Audience: strings.TrimSpace(r.FormValue("audience")), Tags: strings.TrimSpace(r.FormValue("tags")), Source: strings.TrimSpace(r.FormValue("source")), DocumentFile: documentURL, School: strings.TrimSpace(r.FormValue("school")), DocumentType: strings.TrimSpace(r.FormValue("document_type")), Subject: strings.TrimSpace(r.FormValue("subject")), AcademicYear: strings.TrimSpace(r.FormValue("academic_year"))}
	if selectedFound {
		meta.CustomFields = collectCustomFields(r, enabledConfig(selectedCat.FormConfig))
	}
	if err := validatePostMeta(meta); err != nil {
		renderSubmit(h, w, cats, u, err.Error(), oldDraft)
		return
	}
	metaJSON, _ := json.Marshal(meta)
	action := r.FormValue("action")

	if action == "save_draft" {
		d := model.Draft{ID: draftID, CategoryID: cid, Title: r.FormValue("title"), Content: r.FormValue("content"), MetaJSON: string(metaJSON), AuthorID: u.ID}
		id, err := h.content.SaveDraft(d)
		if err != nil {
			renderSubmit(h, w, cats, u, err.Error(), oldDraft)
			return
		}
		http.Redirect(w, r, "/dang-bai?draft="+strconv.Itoa(id)+"&saved=1", 303)
		return
	}

	cat, found := findCategory(cats, cid)
	if !found || !categoryVisibleTo(cat, &u) {
		renderSubmit(h, w, cats, u, "Bạn không có quyền đăng vào chuyên mục này", oldDraft)
		return
	}
	if !u.IsAdmin && audienceScopes(cat.AudienceScope)["same_school"] && (!isStudentViewer(&u) || strings.TrimSpace(u.School) == "") {
		renderSubmit(h, w, cats, u, "Confession chỉ dành cho sinh viên đã cập nhật trường trong hồ sơ", oldDraft)
		return
	}
	if meta.DocumentFile != "" && !documentFormatAllowed(meta.DocumentFile, cat.DocumentFormats) {
		renderSubmit(h, w, cats, u, "File tài liệu hiện tại không thuộc định dạng được phép của chuyên mục này", oldDraft)
		return
	}
	cfg := enabledConfig(cat.FormConfig)
	if err := validateCustomFields(meta, cfg); err != nil {
		renderSubmit(h, w, cats, u, err.Error(), oldDraft)
		return
	}
	checks := map[string]string{"title": r.FormValue("title"), "content": r.FormValue("content"), "deadline": meta.Deadline, "website": meta.Website, "fanpage": meta.Fanpage, "recruitment_content": meta.RecruitmentContent, "cv_email": meta.CVEmail, "contact_name": meta.ContactName, "contact_phone": meta.ContactPhone, "organization": meta.Organization, "location": meta.Location, "salary_range": meta.SalaryRange, "application_link": meta.ApplicationLink, "event_time": meta.EventTime, "audience": meta.Audience, "tags": meta.Tags, "source": meta.Source, "school": meta.School, "document_type": meta.DocumentType, "subject": meta.Subject, "academic_year": meta.AcademicYear}
	for key, val := range checks {
		if required(cfg, key) && strings.TrimSpace(val) == "" {
			renderSubmit(h, w, cats, u, "Vui lòng nhập trường bắt buộc: "+key, oldDraft)
			return
		}
	}
	if required(cfg, "document_file") && meta.DocumentFile == "" {
		renderSubmit(h, w, cats, u, "Vui lòng chọn File tài liệu học tập", oldDraft)
		return
	}
	if required(cfg, "image") && meta.ImageURL == "" {
		renderSubmit(h, w, cats, u, "Vui lòng chọn Ảnh bài viết", oldDraft)
		return
	}
	if required(cfg, "company_logo") && meta.CompanyLogo == "" {
		renderSubmit(h, w, cats, u, "Vui lòng chọn Logo công ty", oldDraft)
		return
	}
	if required(cfg, "positions") && len(meta.Positions) == 0 {
		renderSubmit(h, w, cats, u, "Vui lòng thêm ít nhất một vị trí tuyển", oldDraft)
		return
	}
	post := model.Post{Title: r.FormValue("title"), Content: r.FormValue("content"), CategoryID: cid, MetaJSON: string(metaJSON), AuthorID: u.ID}
	if err := h.content.SavePost(post); err != nil {
		renderSubmit(h, w, cats, u, err.Error(), oldDraft)
		return
	}
	h.syncFilterValues(cat, meta)
	if draftID > 0 {
		_ = h.content.DeleteDraft(draftID, u.ID)
	}
	http.Redirect(w, r, "/?category="+cat.Slug, 303)
}

func (h *Handler) EditPost(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login?next=/ho-so", 303)
		return
	}
	if rejectRestricted(w, r, u) {
		return
	}
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if r.Method == http.MethodPost {
		id, _ = strconv.Atoi(r.FormValue("post_id"))
	}
	post, err := h.content.GetPost(id)
	if err != nil || (post.AuthorID != u.ID && !u.IsAdmin) {
		http.Error(w, "Bạn không có quyền sửa bài viết này", http.StatusForbidden)
		return
	}
	allCats, _ := h.content.ListCategories()
	cats := filterAccessibleCategories(allCats, &u)
	var oldMeta model.PostMeta
	_ = json.Unmarshal([]byte(post.MetaJSON), &oldMeta)
	renderEdit := func(message string, meta model.PostMeta) {
		draft := model.Draft{CategoryID: post.CategoryID, Title: post.Title, Content: post.Content, MetaJSON: post.MetaJSON, AuthorID: u.ID}
		data := model.PageData{Title: "Sửa bài viết", Categories: cats, IsLoggedIn: true, UserName: u.Name, CurrentUser: &u, Message: message, Draft: &draft, DraftMeta: meta, IsEditing: true, EditPostID: post.ID, CustomFieldValuesJSON: customFieldValuesJSON(meta)}
		h.render(w, "submit.html", data)
	}
	if r.Method == http.MethodGet {
		renderEdit("", oldMeta)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		renderEdit("Dữ liệu tải lên quá lớn hoặc không hợp lệ", oldMeta)
		return
	}
	cid, _ := strconv.Atoi(r.FormValue("category_id"))
	selectedCat, selectedFound := findCategory(cats, cid)
	allowedDocFormats := "pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip"
	if selectedFound {
		allowedDocFormats = selectedCat.DocumentFormats
	}
	imageURL, err := saveUpload(r, "image")
	if err != nil {
		renderEdit(err.Error(), oldMeta)
		return
	}
	logoURL, err := saveUpload(r, "company_logo")
	if err != nil {
		renderEdit(err.Error(), oldMeta)
		return
	}
	documentURL, docErr := saveDocumentUpload(r, "document_file", allowedDocFormats)
	if docErr != nil {
		renderEdit(docErr.Error(), oldMeta)
		return
	}
	if imageURL == "" {
		imageURL = oldMeta.ImageURL
	}
	if logoURL == "" {
		logoURL = oldMeta.CompanyLogo
	}
	if documentURL == "" {
		documentURL = oldMeta.DocumentFile
	}
	meta := model.PostMeta{ImageURL: imageURL, Deadline: strings.TrimSpace(r.FormValue("deadline")), CompanyLogo: logoURL, Website: strings.TrimSpace(r.FormValue("website")), Fanpage: strings.TrimSpace(r.FormValue("fanpage")), RecruitmentContent: strings.TrimSpace(r.FormValue("recruitment_content")), CVEmail: strings.TrimSpace(r.FormValue("cv_email")), Positions: collectPositions(r), ContactName: strings.TrimSpace(r.FormValue("contact_name")), ContactPhone: strings.TrimSpace(r.FormValue("contact_phone")), Organization: strings.TrimSpace(r.FormValue("organization")), Location: strings.TrimSpace(r.FormValue("location")), SalaryRange: strings.TrimSpace(r.FormValue("salary_range")), ApplicationLink: strings.TrimSpace(r.FormValue("application_link")), EventTime: strings.TrimSpace(r.FormValue("event_time")), Audience: strings.TrimSpace(r.FormValue("audience")), Tags: strings.TrimSpace(r.FormValue("tags")), Source: strings.TrimSpace(r.FormValue("source")), DocumentFile: documentURL, School: strings.TrimSpace(r.FormValue("school")), DocumentType: strings.TrimSpace(r.FormValue("document_type")), Subject: strings.TrimSpace(r.FormValue("subject")), AcademicYear: strings.TrimSpace(r.FormValue("academic_year"))}
	if selectedFound {
		meta.CustomFields = collectCustomFields(r, enabledConfig(selectedCat.FormConfig))
	}
	if err := validatePostMeta(meta); err != nil {
		renderEdit(err.Error(), meta)
		return
	}
	cat, found := findCategory(cats, cid)
	if !found || !categoryVisibleTo(cat, &u) {
		renderEdit("Bạn không có quyền đăng vào chuyên mục này", meta)
		return
	}
	if !u.IsAdmin && audienceScopes(cat.AudienceScope)["same_school"] && (!isStudentViewer(&u) || strings.TrimSpace(u.School) == "") {
		renderEdit("Confession chỉ dành cho sinh viên đã cập nhật trường trong hồ sơ", meta)
		return
	}
	if meta.DocumentFile != "" && !documentFormatAllowed(meta.DocumentFile, cat.DocumentFormats) {
		renderEdit("File tài liệu hiện tại không thuộc định dạng được phép của chuyên mục này", meta)
		return
	}
	cfg := enabledConfig(cat.FormConfig)
	if err := validateCustomFields(meta, cfg); err != nil {
		renderEdit(err.Error(), meta)
		return
	}
	checks := map[string]string{"title": r.FormValue("title"), "content": r.FormValue("content"), "deadline": meta.Deadline, "website": meta.Website, "fanpage": meta.Fanpage, "recruitment_content": meta.RecruitmentContent, "cv_email": meta.CVEmail, "contact_name": meta.ContactName, "contact_phone": meta.ContactPhone, "organization": meta.Organization, "location": meta.Location, "salary_range": meta.SalaryRange, "application_link": meta.ApplicationLink, "event_time": meta.EventTime, "audience": meta.Audience, "tags": meta.Tags, "source": meta.Source, "school": meta.School, "document_type": meta.DocumentType, "subject": meta.Subject, "academic_year": meta.AcademicYear}
	for key, val := range checks {
		if required(cfg, key) && strings.TrimSpace(val) == "" {
			renderEdit("Vui lòng nhập trường bắt buộc: "+key, meta)
			return
		}
	}
	if required(cfg, "document_file") && meta.DocumentFile == "" {
		renderEdit("Vui lòng chọn File tài liệu học tập", meta)
		return
	}
	if required(cfg, "image") && meta.ImageURL == "" {
		renderEdit("Vui lòng chọn Ảnh bài viết", meta)
		return
	}
	if required(cfg, "company_logo") && meta.CompanyLogo == "" {
		renderEdit("Vui lòng chọn Logo công ty", meta)
		return
	}
	if required(cfg, "positions") && len(meta.Positions) == 0 {
		renderEdit("Vui lòng thêm ít nhất một vị trí tuyển", meta)
		return
	}
	metaJSON, _ := json.Marshal(meta)
	post.Title = r.FormValue("title")
	post.Content = r.FormValue("content")
	post.CategoryID = cid
	post.MetaJSON = string(metaJSON)
	post.AuthorID = u.ID
	if err := h.content.SavePost(post); err != nil {
		renderEdit(err.Error(), meta)
		return
	}
	h.syncFilterValues(cat, meta)
	http.Redirect(w, r, "/bai-viet/"+strconv.Itoa(post.ID), 303)
}

func (h *Handler) DeleteOwnPost(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login", 303)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	post, err := h.content.GetPost(id)
	if err != nil || (post.AuthorID != u.ID && !u.IsAdmin) {
		http.Error(w, "Bạn không có quyền xóa bài viết này", http.StatusForbidden)
		return
	}
	if err := h.content.DeletePost(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/ho-so?msg="+urlMessage("Đã xóa bài viết"), 303)
}

func (h *Handler) MyDrafts(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login?next=/bai-da-luu", 303)
		return
	}
	var drafts []model.Draft
	var savedPosts []model.Post
	var allCats []model.Category
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); drafts, _ = h.content.ListDrafts(u.ID) }()
	go func() { defer wg.Done(); savedPosts, _ = h.content.ListSavedPosts(u.ID) }()
	go func() { defer wg.Done(); allCats, _ = h.content.ListCategories() }()
	wg.Wait()

	authors := h.loadAuthors(savedPosts)
	visible := make([]model.Post, 0, len(savedPosts))
	metaByID := make(map[int]model.PostMeta, len(savedPosts))
	for _, post := range savedPosts {
		if !h.postVisibleTo(post, allCats, &u, authors[post.AuthorID]) {
			continue
		}
		var meta model.PostMeta
		_ = json.Unmarshal([]byte(post.MetaJSON), &meta)
		visible = append(visible, post)
		metaByID[post.ID] = meta
	}
	statsByID, _ := h.content.GetPostStatsBatch(postIDs(visible))
	savedCards := make([]model.PostCard, 0, len(visible))
	for _, post := range visible {
		savedCards = append(savedCards, model.PostCard{Post: post, Meta: metaByID[post.ID], Stats: statsByID[post.ID], Saved: true, Author: authors[post.AuthorID]})
	}
	h.render(w, "drafts.html", model.PageData{Title: "Bài đã lưu", Drafts: drafts, SavedPosts: savedCards, IsLoggedIn: true, UserName: u.Name, CurrentUser: &u})
}

func (h *Handler) DeleteDraft(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login", 303)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if err := h.content.DeleteDraft(id, u.ID); err != nil {
		http.Redirect(w, r, "/bai-da-luu?msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/bai-da-luu", 303)
}

func (h *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	u, ok := h.sessions.CurrentUser(r)
	if !ok || !u.IsAdmin {
		http.Redirect(w, r, "/login?next=/admin", 303)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	filter := strings.TrimSpace(r.URL.Query().Get("filter"))
	reportStatus := strings.TrimSpace(r.URL.Query().Get("report_status"))
	adminUserQuery := strings.TrimSpace(r.URL.Query().Get("user_q"))
	filterCategoryID, _ := strconv.Atoi(r.URL.Query().Get("filter_category"))
	filterField := strings.TrimSpace(r.URL.Query().Get("filter_field"))
	if reportStatus == "" {
		reportStatus = "pending"
	}

	var posts []model.Post
	var reports, allReports []model.PostReport
	var cats []model.Category
	var ads []model.Advertisement
	var verificationRequests []model.VerificationRequest
	var adminUsers []model.User
	var filterValues []model.FilterValue
	var wg sync.WaitGroup
	run := func(fn func()) { wg.Add(1); go func() { defer wg.Done(); fn() }() }
	run(func() { posts, _ = h.content.ListPosts(q, category) })
	run(func() { reports, _ = h.content.ListPostReports(reportStatus) })
	run(func() { cats, _ = h.content.ListCategories() })
	run(func() { ads, _ = h.content.ListAds("", false) })
	run(func() { verificationRequests, _ = h.content.ListVerificationRequests("all") })
	run(func() { adminUsers, _ = h.auth.SearchUsers(adminUserQuery, 20) })
	run(func() { filterValues, _ = h.content.ListFilterValues(filterCategoryID, filterField, "") })
	if filter == "reported" {
		run(func() { allReports, _ = h.content.ListPostReports("all") })
	}
	wg.Wait()

	if filter == "reported" {
		reported := map[int]bool{}
		for _, item := range allReports {
			reported[item.PostID] = true
		}
		filtered := posts[:0]
		for _, post := range posts {
			if reported[post.ID] {
				filtered = append(filtered, post)
			}
		}
		posts = filtered
	}
	for i := range filterValues {
		filterValues[i].FieldLabel = filterValues[i].FieldKey
		for _, c := range cats {
			if c.ID == filterValues[i].CategoryID {
				filterValues[i].FieldLabel = fieldLabelFromConfig(c, filterValues[i].FieldKey)
				break
			}
		}
	}
	h.render(w, "admin.html", model.PageData{Title: "Quản trị", Posts: posts, Categories: cats, Ads: ads, IsAdmin: true, CurrentUser: &u, IsLoggedIn: true, UserName: u.Name, Message: r.URL.Query().Get("msg"), Reports: reports, AdminQuery: q, AdminCategory: category, AdminFilter: filter, ReportStatus: reportStatus, VerificationRequests: verificationRequests, AdminUserQuery: adminUserQuery, AdminUsers: adminUsers, FilterValues: filterValues, AdminFilterCategoryID: filterCategoryID, AdminFilterField: filterField})
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	u, ok := h.sessions.CurrentUser(r)
	if !ok || !u.IsAdmin {
		http.Redirect(w, r, "/login?next=/admin", 303)
		return false
	}
	return true
}

func wantsJSON(r *http.Request) bool {
	return strings.Contains(strings.ToLower(r.Header.Get("Accept")), "application/json") || strings.EqualFold(r.Header.Get("X-Requested-With"), "XMLHttpRequest")
}

func writeAdminJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// parseMutationForm parses both ordinary URL-encoded forms and multipart FormData.
// Several admin actions are submitted with JavaScript FormData, so ParseForm alone
// would leave multipart fields out of r.Form and make a write appear successful
// while persisting default/empty values.
func parseMutationForm(r *http.Request) error {
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return r.ParseMultipartForm(4 << 20)
	}
	return r.ParseForm()
}

func hasCategoryConfigPayload(values url.Values) bool {
	for key := range values {
		if strings.HasPrefix(key, "label_") ||
			strings.HasPrefix(key, "enabled_") ||
			strings.HasPrefix(key, "required_") ||
			strings.HasPrefix(key, "filterable_") ||
			strings.HasPrefix(key, "suggest_") ||
			strings.HasPrefix(key, "allow_custom_") ||
			strings.HasPrefix(key, "order_") {
			return true
		}
	}
	return false
}

func (h *Handler) SavePost(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	cid, _ := strconv.Atoi(r.FormValue("category_id"))
	p := model.Post{ID: id, Title: r.FormValue("title"), Summary: r.FormValue("summary"), Content: r.FormValue("content"), CategoryID: cid}
	if id > 0 {
		if old, err := h.content.GetPost(id); err == nil {
			p.MetaJSON = old.MetaJSON
			p.AuthorID = old.AuthorID
			p.IsPinned = old.IsPinned
		}
	}
	if err := h.content.SavePost(p); err != nil {
		http.Redirect(w, r, "/admin?msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/admin?msg=Đã+lưu+bài+viết", 303)
}
func (h *Handler) PinPost(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	pinned := r.FormValue("pinned") == "1"
	if err := h.content.SetPostPinned(id, pinned); err != nil {
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		http.Redirect(w, r, "/admin?msg="+urlMessage(err.Error()), 303)
		return
	}
	if wantsJSON(r) {
		writeAdminJSON(w, 200, map[string]any{"ok": true, "pinned": pinned, "message": "Đã lưu trạng thái ghim"})
		return
	}
	msg := "Đã+bỏ+ghim+bài+viết"
	if pinned {
		msg = "Đã+ghim+bài+viết"
	}
	http.Redirect(w, r, "/admin?msg="+msg, 303)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if err := h.content.DeletePost(id); err != nil {
		http.Redirect(w, r, "/admin?msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/admin?msg=Đã+xóa+bài+viết", 303)
}
func (h *Handler) EducationOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"schools": reference.VietnamSchools,
		"majors":  reference.CommonMajors,
		"version": "VN-EDU-2026-09",
	})
}

func (h *Handler) FilterOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	cid, _ := strconv.Atoi(r.URL.Query().Get("category_id"))
	field := strings.TrimSpace(r.URL.Query().Get("field"))
	if cid <= 0 || field == "" {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "values": []string{}})
		return
	}
	vals, err := h.content.ListApprovedFilterValues(cid, field)
	if err != nil {
		writeAdminJSON(w, 500, map[string]any{"ok": false, "message": "Không thể đọc dữ liệu gợi ý"})
		return
	}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		out = append(out, v.Value)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "values": out})
}
func (h *Handler) AdminFilterValueAction(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	action := strings.TrimSpace(r.FormValue("action"))
	value := strings.TrimSpace(r.FormValue("value"))
	if err := h.content.ReviewFilterValue(id, action, value); err != nil {
		http.Redirect(w, r, "/admin?tab=filter-data&msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/admin?tab=filter-data&msg="+urlMessage("Đã cập nhật dữ liệu bộ lọc"), 303)
}

func (h *Handler) SaveCategory(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	_ = r.ParseForm()
	cfg := service.BuildFieldConfig(r.Form)
	showCompanyCard := r.FormValue("show_company_card") != ""
	companyCardStyle := r.FormValue("company_card_style")
	postCardStyle := r.FormValue("post_card_style")
	documentFormats := strings.Join(r.Form["document_formats"], ",")
	documentFormats = service.NormalizeDocumentFormats(documentFormats)
	audienceScope := strings.Join(r.Form["audience_scope"], ",")
	if strings.TrimSpace(audienceScope) == "" {
		audienceScope = "public"
	}
	if err := h.content.SaveCategory(model.Category{Name: r.FormValue("name"), Slug: r.FormValue("slug"), FormConfig: cfg, ShowCompanyCard: showCompanyCard, CompanyCardStyle: companyCardStyle, AudienceScope: audienceScope, PostCardStyle: postCardStyle, DocumentFormats: documentFormats}); err != nil {
		http.Redirect(w, r, "/admin?msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/admin?msg=Đã+thêm+chuyên+mục+và+cấu+hình+form", 303)
}
func (h *Handler) SaveCategoryConfig(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	if err := parseMutationForm(r); err != nil {
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": "Dữ liệu cấu hình không hợp lệ: " + err.Error()})
		return
	}

	rawID := strings.TrimSpace(r.FormValue("id"))
	if rawID == "" {
		rawID = strings.TrimSpace(r.FormValue("category_id"))
	}
	if rawID == "" {
		rawID = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	categorySlug := strings.TrimSpace(r.FormValue("category_slug"))
	if categorySlug == "" {
		categorySlug = strings.TrimSpace(r.URL.Query().Get("category_slug"))
	}

	id, parseErr := strconv.Atoi(rawID)
	cats, listErr := h.content.ListCategories()
	if listErr != nil {
		if wantsJSON(r) {
			writeAdminJSON(w, 500, map[string]any{"ok": false, "message": "Không thể đọc danh sách chuyên mục"})
			return
		}
		http.Redirect(w, r, "/admin?tab=categories&msg="+urlMessage("Không thể đọc danh sách chuyên mục"), 303)
		return
	}

	var cat model.Category
	found := false
	if parseErr == nil && id > 0 {
		cat, found = findCategory(cats, id)
	}
	if !found && categorySlug != "" {
		for _, candidate := range cats {
			if strings.EqualFold(strings.TrimSpace(candidate.Slug), categorySlug) {
				cat = candidate
				id = candidate.ID
				found = true
				break
			}
		}
	}
	if !found || id <= 0 {
		msg := "Không xác định được chuyên mục cần lưu. Hãy tải lại trang quản trị rồi mở Cấu hình form."
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": msg})
			return
		}
		http.Redirect(w, r, "/admin?tab=categories&msg="+urlMessage(msg), 303)
		return
	}

	// Never report success if the form-builder payload is missing.
	if !hasCategoryConfigPayload(r.Form) {
		msg := "Không nhận được dữ liệu các trường form. Vui lòng tải lại trang rồi lưu lại."
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": msg})
			return
		}
		http.Redirect(w, r, "/admin?tab=categories&msg="+urlMessage(msg), 303)
		return
	}

	cfg := service.BuildFieldConfig(r.Form)

	// V1.24.25: "Cấu hình form" chỉ quản lý các field của form đăng bài.
	// Các thuộc tính chuyên mục (tên, đối tượng xem, kiểu thẻ...) được chỉnh
	// trực tiếp ngoài danh sách và phải được giữ nguyên khi lưu form.
	updated := model.Category{
		ID:               id,
		Slug:             cat.Slug,
		FormConfig:       cfg,
		ShowCompanyCard:  cat.ShowCompanyCard,
		CompanyCardStyle: cat.CompanyCardStyle,
		AudienceScope:    cat.AudienceScope,
		PostCardStyle:    cat.PostCardStyle,
		DocumentFormats:  cat.DocumentFormats,
	}
	if err := h.content.UpdateCategoryConfig(updated); err != nil {
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		http.Redirect(w, r, "/admin?tab=categories&msg="+urlMessage(err.Error()), 303)
		return
	}

	// Read back from the repository before returning success.
	savedCategories, err := h.content.ListCategories()
	if err != nil {
		if wantsJSON(r) {
			writeAdminJSON(w, 500, map[string]any{"ok": false, "message": "Đã ghi dữ liệu nhưng không thể đọc lại chuyên mục để xác nhận"})
			return
		}
		http.Redirect(w, r, "/admin?tab=categories&msg="+urlMessage("Đã ghi dữ liệu nhưng không thể đọc lại để xác nhận"), 303)
		return
	}
	saved, ok := findCategory(savedCategories, id)
	if !ok {
		if wantsJSON(r) {
			writeAdminJSON(w, 500, map[string]any{"ok": false, "message": "Không tìm thấy chuyên mục sau khi lưu"})
			return
		}
		http.Redirect(w, r, "/admin?tab=categories&msg="+urlMessage("Không tìm thấy chuyên mục sau khi lưu"), 303)
		return
	}
	expectedConfig := service.NormalizeConfig(cat.Slug, cfg)
	if saved.FormConfig != expectedConfig {
		msg := "CSDL chưa phản ánh cấu hình vừa lưu. Vui lòng thử lại."
		if wantsJSON(r) {
			writeAdminJSON(w, 500, map[string]any{"ok": false, "message": msg})
			return
		}
		http.Redirect(w, r, "/admin?tab=categories&msg="+urlMessage(msg), 303)
		return
	}

	if wantsJSON(r) {
		writeAdminJSON(w, 200, map[string]any{
			"ok":       true,
			"message":  "Đã lưu và xác nhận dữ liệu chuyên mục",
			"category": saved,
		})
		return
	}
	http.Redirect(w, r, "/admin?tab=categories&msg="+urlMessage("Đã lưu và xác nhận dữ liệu chuyên mục"), 303)
}

func (h *Handler) UpdateCategoryMeta(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	if err := parseMutationForm(r); err != nil {
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": "Dữ liệu chuyên mục không hợp lệ: " + err.Error()})
		return
	}
	id, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("id")))
	if id <= 0 {
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": "Không xác định được chuyên mục"})
		return
	}
	cats, err := h.content.ListCategories()
	if err != nil {
		writeAdminJSON(w, 500, map[string]any{"ok": false, "message": "Không thể đọc chuyên mục"})
		return
	}
	cat, found := findCategory(cats, id)
	if !found {
		writeAdminJSON(w, 404, map[string]any{"ok": false, "message": "Chuyên mục không còn tồn tại"})
		return
	}

	field := strings.TrimSpace(r.FormValue("field"))
	switch field {
	case "name":
		cat.Name = strings.TrimSpace(r.FormValue("name"))
	case "audience":
		cat.AudienceScope = strings.Join(r.Form["audience_scope"], ",")
		if strings.TrimSpace(cat.AudienceScope) == "" {
			cat.AudienceScope = "public"
		}
	case "post_card_style":
		cat.PostCardStyle = strings.TrimSpace(r.FormValue("post_card_style"))
		if formats, ok := r.Form["document_formats"]; ok {
			cat.DocumentFormats = service.NormalizeDocumentFormats(strings.Join(formats, ","))
		}
	case "company_card":
		cat.ShowCompanyCard = r.FormValue("show_company_card") != ""
		cat.CompanyCardStyle = strings.TrimSpace(r.FormValue("company_card_style"))
	default:
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": "Thuộc tính chuyên mục không hợp lệ"})
		return
	}

	if err := h.content.UpdateCategoryMeta(cat); err != nil {
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
		return
	}
	updatedCats, err := h.content.ListCategories()
	if err != nil {
		writeAdminJSON(w, 500, map[string]any{"ok": false, "message": "Đã lưu nhưng không thể đọc lại dữ liệu"})
		return
	}
	saved, ok := findCategory(updatedCats, id)
	if !ok {
		writeAdminJSON(w, 500, map[string]any{"ok": false, "message": "Không tìm thấy chuyên mục sau khi lưu"})
		return
	}
	writeAdminJSON(w, 200, map[string]any{"ok": true, "message": "Đã cập nhật chuyên mục", "category": saved, "field": field})
}

func (h *Handler) ReorderCategories(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	var payload struct {
		IDs []int `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": "Dữ liệu thứ tự không hợp lệ"})
		return
	}
	if err := h.content.ReorderCategories(payload.IDs); err != nil {
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
		return
	}
	cats, _ := h.content.ListCategories()
	writeAdminJSON(w, 200, map[string]any{"ok": true, "message": "Đã lưu thứ tự chuyên mục", "categories": cats})
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if err := h.content.DeleteCategory(id); err != nil {
		http.Redirect(w, r, "/admin?msg=Không+thể+xóa+chuyên+mục+đang+có+bài+viết", 303)
		return
	}
	http.Redirect(w, r, "/admin?msg=Đã+xóa+chuyên+mục", 303)
}
func (h *Handler) SaveAd(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": "Dữ liệu quảng cáo không hợp lệ"})
			return
		}
		http.Redirect(w, r, "/admin?msg=Du+lieu+quang+cao+khong+hop+le", 303)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	order, _ := strconv.Atoi(r.FormValue("sort_order"))
	var oldAd model.Advertisement
	if id > 0 {
		oldAd, _ = h.content.GetAd(id)
	}
	imageURL, err := saveUpload(r, "ad_image")
	if err != nil {
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		http.Redirect(w, r, "/admin?msg="+urlMessage(err.Error()), 303)
		return
	}
	newImageUploaded := imageURL != ""
	if id > 0 && imageURL == "" {
		imageURL = oldAd.ImageURL
	}
	ad := model.Advertisement{ID: id, Title: r.FormValue("title"), Description: r.FormValue("description"), ImageURL: imageURL, LinkURL: r.FormValue("link_url"), Position: r.FormValue("position"), Active: r.FormValue("active") == "on", SortOrder: order}
	if err := h.content.SaveAd(ad); err != nil {
		if newImageUploaded {
			_ = deleteStoredUpload(imageURL)
		}
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		http.Redirect(w, r, "/admin?msg="+urlMessage(err.Error()), 303)
		return
	}
	if newImageUploaded && oldAd.ImageURL != "" && oldAd.ImageURL != imageURL {
		_ = deleteStoredUpload(oldAd.ImageURL)
	}
	if wantsJSON(r) {
		ads, _ := h.content.ListAds("", false)
		writeAdminJSON(w, 200, map[string]any{"ok": true, "message": "Đã lưu quảng cáo", "ads": ads})
		return
	}
	http.Redirect(w, r, "/admin?msg=Da+luu+quang+cao", 303)
}

func (h *Handler) ReorderAds(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	var payload struct {
		IDs []int `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": "Thứ tự quảng cáo không hợp lệ"})
		return
	}
	if err := h.content.ReorderAds(payload.IDs); err != nil {
		writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
		return
	}
	ads, _ := h.content.ListAds("", false)
	writeAdminJSON(w, 200, map[string]any{"ok": true, "message": "Đã tự động lưu thứ tự quảng cáo", "ads": ads})
}

func (h *Handler) DeleteAd(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	oldAd, _ := h.content.GetAd(id)
	if err := h.content.DeleteAd(id); err != nil {
		if wantsJSON(r) {
			writeAdminJSON(w, 400, map[string]any{"ok": false, "message": err.Error()})
			return
		}
		http.Redirect(w, r, "/admin?msg="+urlMessage(err.Error()), 303)
		return
	}
	if oldAd.ImageURL != "" {
		_ = deleteStoredUpload(oldAd.ImageURL)
	}
	if wantsJSON(r) {
		ads, _ := h.content.ListAds("", false)
		writeAdminJSON(w, 200, map[string]any{"ok": true, "message": "Đã xóa quảng cáo", "ads": ads})
		return
	}
	http.Redirect(w, r, "/admin?msg="+urlMessage("Đã xóa quảng cáo"), 303)
}

func (h *Handler) ReportPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	u, ok := h.sessions.CurrentUser(r)
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	next := safeNext(r.FormValue("next"), "/")
	if !ok {
		http.Redirect(w, r, "/login?next="+next, 303)
		return
	}
	if rejectRestricted(w, r, u) {
		return
	}
	if !h.userCanAccessPost(postID, &u) {
		http.Error(w, "Bạn không có quyền truy cập bài viết này", http.StatusForbidden)
		return
	}
	if post, err := h.content.GetPost(postID); err == nil && post.AuthorID == u.ID {
		http.Redirect(w, r, next+"?msg="+urlMessage("Bạn không thể báo cáo bài viết của chính mình"), 303)
		return
	}
	if err := h.content.ReportPost(postID, u.ID, r.FormValue("reason")); err != nil {
		http.Redirect(w, r, next+"?msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, next+"?msg="+urlMessage("Đã gửi báo cáo cho quản trị viên"), 303)
}

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	u, ok := h.sessions.CurrentUser(r)
	if !ok {
		http.Redirect(w, r, "/login?next=/", 303)
		return
	}
	blockedID, _ := strconv.Atoi(r.FormValue("user_id"))
	if err := h.content.BlockUser(u.ID, blockedID); err != nil {
		http.Redirect(w, r, safeNext(r.FormValue("next"), "/")+"?msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/?msg="+urlMessage("Đã chặn tài khoản. Các bài viết của tài khoản này sẽ được ẩn khỏi bảng tin của bạn."), 303)
}

func (h *Handler) AdminReportAction(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	action := strings.TrimSpace(r.FormValue("action"))

	var err error
	switch action {
	case "resolve":
		err = h.content.UpdatePostReportStatus(id, "resolved")
	case "dismiss":
		err = h.content.UpdatePostReportStatus(id, "dismissed")
	case "delete":
		if deleteErr := h.content.DeletePost(postID); deleteErr != nil {
			err = deleteErr
		} else {
			err = h.content.UpdatePostReportStatus(id, "resolved")
		}
	default:
		err = h.content.UpdatePostReportStatus(id, "pending")
	}
	if err != nil {
		http.Redirect(w, r, "/admin?tab=moderation&msg="+urlMessage(err.Error()), 303)
		return
	}
	http.Redirect(w, r, "/admin?tab=moderation&msg="+urlMessage("Đã cập nhật kiểm duyệt"), 303)
}

func (h *Handler) AdminUserAction(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) || r.Method != http.MethodPost {
		return
	}
	admin, ok := h.sessions.CurrentUser(r)
	if !ok {
		return
	}
	targetID, _ := strconv.Atoi(r.FormValue("user_id"))
	action := strings.TrimSpace(r.FormValue("action"))
	query := strings.TrimSpace(r.FormValue("user_q"))
	redirect := "/admin?tab=moderation"
	if query != "" {
		redirect += "&user_q=" + url.QueryEscape(query)
	}
	if targetID <= 0 {
		http.Redirect(w, r, redirect+"&msg="+url.QueryEscape("Tài khoản không hợp lệ"), 303)
		return
	}
	target, err := h.auth.GetUserByID(targetID)
	if err != nil {
		http.Redirect(w, r, redirect+"&msg="+url.QueryEscape("Không tìm thấy tài khoản"), 303)
		return
	}
	if target.ID == admin.ID && (action == "delete" || action == "restrict" || strings.HasPrefix(action, "lock") || action == "revoke_admin") {
		http.Redirect(w, r, redirect+"&msg="+url.QueryEscape("Không thể tự khóa, hạn chế, xóa hoặc thu hồi quyền quản trị của chính mình"), 303)
		return
	}
	message := "Đã cập nhật tài khoản"
	switch action {
	case "grant_admin":
		err = h.auth.SetUserAdmin(targetID, true)
		message = "Đã cấp quyền quản trị"
	case "revoke_admin":
		count, _ := h.auth.CountAdmins()
		if target.IsAdmin && count <= 1 {
			err = errors.New("Không thể thu hồi quản trị viên cuối cùng")
		} else {
			err = h.auth.SetUserAdmin(targetID, false)
			message = "Đã thu hồi quyền quản trị"
		}
	case "restrict":
		err = h.auth.SetUserAccountStatus(targetID, "restricted", "")
		message = "Đã hạn chế tài khoản"
	case "activate", "unlock":
		err = h.auth.SetUserAccountStatus(targetID, "active", "")
		message = "Đã mở hạn chế / mở khóa tài khoản"
	case "lock_24h":
		err = h.auth.SetUserAccountStatus(targetID, "locked", time.Now().Add(24*time.Hour).Format(time.RFC3339))
		message = "Đã tạm khóa tài khoản 24 giờ"
	case "lock_7d":
		err = h.auth.SetUserAccountStatus(targetID, "locked", time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
		message = "Đã tạm khóa tài khoản 7 ngày"
	case "delete":
		if target.IsAdmin {
			count, _ := h.auth.CountAdmins()
			if count <= 1 {
				err = errors.New("Không thể xóa quản trị viên cuối cùng")
			}
		}
		if err == nil {
			err = h.auth.DeleteUser(targetID)
			message = "Đã xóa tài khoản và các bài viết của tài khoản"
		}
	default:
		err = errors.New("Thao tác không hợp lệ")
	}
	if err != nil {
		message = err.Error()
	}
	http.Redirect(w, r, redirect+"&msg="+url.QueryEscape(message), 303)
}

func urlMessage(s string) string { return strings.ReplaceAll(s, " ", "+") }

var _ = enabled

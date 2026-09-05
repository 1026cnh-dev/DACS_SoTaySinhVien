package repository

import "sotaysinhvien/internal/model"

type ContentRepository interface {
	Close() error
	ListCategories() ([]model.Category, error)
	ListPosts(query, category string) ([]model.Post, error)
	ListPostsLimited(query, category string, limit int) ([]model.Post, error)
	ListPostsByAuthor(authorID int) ([]model.Post, error)
	ListPinnedPosts(category string, limit int) ([]model.Post, error)
	ListTodayPosts(category string, limit int) ([]model.Post, error)
	ListTrendingPosts(limit int) ([]model.Post, error)
	GetPost(id int) (model.Post, error)
	SavePost(post model.Post) error
	SetPostPinned(id int, pinned bool) error
	DeletePost(id int) error
	SaveCategory(category model.Category) error
	UpdateCategoryConfig(category model.Category) error
	UpdateCategoryMeta(category model.Category) error
	ReorderCategories(ids []int) error
	DeleteCategory(id int) error
	UpsertFilterValue(categoryID int, fieldKey, value string, approved bool) error
	ListFilterValues(categoryID int, fieldKey, status string) ([]model.FilterValue, error)
	ReviewFilterValue(id int, action, newValue string) error
	VotePost(postID, userID, value int) error
	GetPostStats(postID int) (model.PostStats, error)
	GetPostStatsBatch(postIDs []int) (map[int]model.PostStats, error)
	GetUserPostVote(postID, userID int) (int, error)
	CreateComment(comment model.Comment) error
	ListComments(postID int) ([]model.Comment, error)
	ListAds(position string, activeOnly bool) ([]model.Advertisement, error)
	GetAd(id int) (model.Advertisement, error)
	SaveAd(ad model.Advertisement) error
	ReorderAds(ids []int) error
	DeleteAd(id int) error
	SaveDraft(draft model.Draft) (int, error)
	GetDraft(id, authorID int) (model.Draft, error)
	ListDrafts(authorID int) ([]model.Draft, error)
	DeleteDraft(id, authorID int) error
	IsPostSaved(postID, userID int) (bool, error)
	ListSavedPostIDs(userID int, postIDs []int) (map[int]bool, error)
	SavePostForUser(postID, userID int) error
	UnsavePostForUser(postID, userID int) error
	ListSavedPosts(userID int) ([]model.Post, error)
	CreatePostReport(report model.PostReport) error
	ListPostReports(status string) ([]model.PostReport, error)
	UpdatePostReportStatus(id int, status string) error
	BlockUser(blockerID, blockedID int) error
	IsUserBlocked(blockerID, blockedID int) (bool, error)
	ListBlockedUserIDs(blockerID int) ([]int, error)
	CreateVerificationRequest(req model.VerificationRequest) error
	GetLatestVerificationRequest(userID int) (model.VerificationRequest, error)
	HasPendingVerificationRequest(userID int, requestType string) (bool, error)
	ListVerificationRequests(status string) ([]model.VerificationRequest, error)
	ResolveVerificationRequest(id int, status string) error

	CreateUser(user model.User) (model.User, error)
	FindUserByLogin(login string) (model.User, error)
	FindUserByEmail(email string) (model.User, error)
	FindUserByPhone(phone string) (model.User, error)
	FindUserByID(id int) (model.User, error)
	FindUsersByIDs(ids []int) (map[int]model.User, error)
	UpdateUserPasswordHash(id int, hash string) error
	UpdateUserProfile(id int, name, username, avatarURL, profileRole, school, major, studentID, phone, employerCompany, employerTaxCode, employerRepresentative, employerWebsite, landlordName, landlordAddress, landlordPhone, landlordLegalInfo string) error
	SetUserVerification(id int, verified bool, verificationType string) error
	SetUserTrustVerified(id int, verified bool) error
	SetUserRoleVerification(id int, verificationType string) error
	SetUserAdmin(id int, isAdmin bool) error
	SearchUsers(query string, limit int) ([]model.User, error)
	SetUserAccountStatus(id int, status, lockedUntil string) error
	DeleteUser(id int) error
	CountAdmins() (int, error)
	UpsertGoogleUser(name, email, googleSub string) (model.User, error)
}

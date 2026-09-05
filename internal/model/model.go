package model

import "time"

type FieldConfig struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Type          string `json:"type,omitempty"`
	Enabled       bool   `json:"enabled"`
	Required      bool   `json:"required"`
	Order         int    `json:"order"`
	Filterable    bool   `json:"filterable"`
	SuggestValues bool   `json:"suggest_values"`
	AllowCustom   bool   `json:"allow_custom"`
}

type FilterValue struct {
	ID           int
	CategoryID   int
	CategoryName string
	FieldKey     string
	FieldLabel   string
	Value        string
	Normalized   string
	Status       string
	UsageCount   int
	CreatedAt    time.Time
}

type FilterOption struct {
	Value string
	Count int
}

type FilterGroup struct {
	Key      string
	Label    string
	Selected string
	Options  []FilterOption
}

type Category struct {
	ID               int
	Name             string
	Slug             string
	FormConfig       string
	ShowCompanyCard  bool
	CompanyCardStyle string
	SortOrder        int
	AudienceScope    string
	PostCardStyle    string
	DocumentFormats  string
	NavLocked        bool
	NavAccessHint    string
}

type RecruitmentPosition struct {
	Label       string `json:"label"`
	Name        string `json:"name"`
	JobType     string `json:"job_type"`
	Specialty   string `json:"specialty"`
	Location    string `json:"location"`
	Salary      string `json:"salary"`
	Experience  string `json:"experience"`
	Description string `json:"description"`
	Requirement string `json:"requirement"`
	Benefit     string `json:"benefit"`
}

type PostMeta struct {
	ImageURL           string                `json:"image_url"`
	Deadline           string                `json:"deadline"`
	CompanyLogo        string                `json:"company_logo"`
	Website            string                `json:"website"`
	Fanpage            string                `json:"fanpage"`
	RecruitmentContent string                `json:"recruitment_content"`
	CVEmail            string                `json:"cv_email"`
	Positions          []RecruitmentPosition `json:"positions"`
	ContactName        string                `json:"contact_name"`
	ContactPhone       string                `json:"contact_phone"`
	Organization       string                `json:"organization"`
	Location           string                `json:"location"`
	SalaryRange        string                `json:"salary_range"`
	ApplicationLink    string                `json:"application_link"`
	EventTime          string                `json:"event_time"`
	Audience           string                `json:"audience"`
	Tags               string                `json:"tags"`
	Source             string                `json:"source"`
	DocumentFile       string                `json:"document_file"`
	School             string                `json:"school"`
	DocumentType       string                `json:"document_type"`
	Subject            string                `json:"subject"`
	AcademicYear       string                `json:"academic_year"`
	CustomFields       map[string]string     `json:"custom_fields,omitempty"`
}

type PostCustomField struct {
	Key   string
	Label string
	Value string
}

type Post struct {
	ID           int
	Title        string
	Summary      string
	Content      string
	CategoryID   int
	CategoryName string
	CategorySlug string
	MetaJSON     string
	AuthorID     int
	IsPinned     bool
	CreatedAt    time.Time
}

type User struct {
	ID                     int
	Name                   string
	Email                  string
	Username               string
	PasswordHash           string
	Provider               string
	GoogleSub              string
	AvatarURL              string
	IsAdmin                bool
	IsVerified             bool
	VerificationType       string
	ProfileRole            string
	School                 string
	Major                  string
	StudentID              string
	Phone                  string
	PhoneVerified          bool
	EmployerCompany        string
	EmployerTaxCode        string
	EmployerRepresentative string
	EmployerWebsite        string
	LandlordName           string
	LandlordAddress        string
	LandlordPhone          string
	LandlordLegalInfo      string
	AccountStatus          string
	LockedUntil            string
	CreatedAt              time.Time
}

type VerificationRequest struct {
	ID        int
	UserID    int
	UserName  string
	UserEmail string
	Type      string
	Info      string
	Status    string
	CreatedAt time.Time
}

type PostCard struct {
	Post      Post
	Meta      PostMeta
	Stats     PostStats
	Author    *User
	Saved     bool
	CardStyle string
}

type PostStats struct {
	Upvotes   int
	Downvotes int
	Score     int
	Comments  int
}

type Draft struct {
	ID           int
	CategoryID   int
	CategoryName string
	Title        string
	Content      string
	MetaJSON     string
	AuthorID     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Advertisement struct {
	ID          int
	Title       string
	Description string
	ImageURL    string
	LinkURL     string
	Position    string
	Active      bool
	SortOrder   int
	CreatedAt   time.Time
}

type PostReport struct {
	ID           int
	PostID       int
	PostTitle    string
	PostAuthorID int
	PostAuthor   string
	ReporterID   int
	ReporterName string
	Reason       string
	Status       string
	CreatedAt    time.Time
}

type Comment struct {
	ID        int
	PostID    int
	ParentID  int
	UserID    int
	UserName  string
	Content   string
	CreatedAt time.Time
	Depth     int
	Children  []Comment
}

type PageData struct {
	Title                    string
	Categories               []Category
	Posts                    []Post
	Post                     *Post
	PostMeta                 PostMeta
	PostCategory             *Category
	Category                 string
	CategoryName             string
	IsDocumentCategory       bool
	Query                    string
	ReturnTo                 string
	Message                  string
	CategoryAccessBlocked    bool
	CategoryAccessMessage    string
	CategoryAccessURL        string
	CategoryAccessAction     string
	CustomFieldValuesJSON    string
	PostCustomFields         []PostCustomField
	IsAdmin                  bool
	IsLoggedIn               bool
	UserName                 string
	AuthMode                 string
	GoogleClientID           string
	CurrentUser              *User
	Author                   *User
	Suggested                []PostCard
	Trending                 []PostCard
	Pinned                   []PostCard
	TodayPosts               []PostCard
	HotPosts                 []PostCard
	FeedAds                  []Advertisement
	PostStats                PostStats
	UserPostVote             int
	Comments                 []Comment
	Ads                      []Advertisement
	HomeAds                  []Advertisement
	LeftAd                   *Advertisement
	RightAd                  *Advertisement
	Draft                    *Draft
	DraftMeta                PostMeta
	Drafts                   []Draft
	SavedPosts               []PostCard
	ProfileUser              *User
	ProfilePosts             []PostCard
	IsOwnProfile             bool
	IsOwnPost                bool
	IsEditing                bool
	EditPostID               int
	Reports                  []PostReport
	AdminQuery               string
	AdminCategory            string
	AdminFilter              string
	ReportStatus             string
	BlockedAuthor            bool
	VerificationRequest      *VerificationRequest
	VerificationRequests     []VerificationRequest
	AdminUserQuery           string
	AdminUsers               []User
	FilterGroups             []FilterGroup
	FilterValues             []FilterValue
	AdminFilterCategoryID    int
	AdminFilterField         string
	EducationEmailOK         bool
	PhoneVerificationPending bool
}

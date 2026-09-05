package reference

import (
	"sort"
	"strings"
	"unicode"
)

type School struct {
	Code    string   `json:"code,omitempty"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}

// VietnamSchools is a built-in canonical catalogue for student profiles.
// The list is intentionally stored in application code so profile matching does
// not depend on an external service at runtime. It combines Vietnamese public
// higher-education institutions, major academies and commonly enrolled private/
// international universities. Names use current/common Vietnamese display names.
var VietnamSchools = []School{
	{"DLU", "Trường Đại học Đà Lạt", []string{"Đại học Đà Lạt", "Dalat University"}},
	{"HUST", "Đại học Bách khoa Hà Nội", []string{"Trường Đại học Bách khoa Hà Nội", "Bách khoa Hà Nội"}},
	{"VNU", "Đại học Quốc gia Hà Nội", []string{"ĐHQG Hà Nội", "VNU"}},
	{"VNU-UET", "Trường Đại học Công nghệ - Đại học Quốc gia Hà Nội", []string{"Đại học Công nghệ ĐHQGHN", "UET"}},
	{"VNU-HUS", "Trường Đại học Khoa học Tự nhiên - Đại học Quốc gia Hà Nội", []string{"Đại học Khoa học Tự nhiên ĐHQGHN", "HUS"}},
	{"VNU-USSH", "Trường Đại học Khoa học Xã hội và Nhân văn - Đại học Quốc gia Hà Nội", []string{"Đại học KHXH&NV ĐHQGHN", "USSH Hà Nội"}},
	{"VNU-UL", "Trường Đại học Luật - Đại học Quốc gia Hà Nội", []string{"Đại học Luật ĐHQGHN", "VNU University of Law"}},
	{"VNU-UEB", "Trường Đại học Kinh tế - Đại học Quốc gia Hà Nội", []string{"Đại học Kinh tế ĐHQGHN", "UEB"}},
	{"VNU-ULIS", "Trường Đại học Ngoại ngữ - Đại học Quốc gia Hà Nội", []string{"Đại học Ngoại ngữ ĐHQGHN", "ULIS"}},
	{"VNU-UEd", "Trường Đại học Giáo dục - Đại học Quốc gia Hà Nội", []string{"Đại học Giáo dục ĐHQGHN"}},
	{"VNU-UMP", "Trường Đại học Y Dược - Đại học Quốc gia Hà Nội", []string{"Đại học Y Dược ĐHQGHN"}},
	{"VNU-IS", "Trường Quốc tế - Đại học Quốc gia Hà Nội", []string{"Trường Quốc tế ĐHQGHN", "VNU-IS"}},
	{"NEU", "Đại học Kinh tế Quốc dân", []string{"Trường Đại học Kinh tế Quốc dân", "NEU"}},
	{"FTU", "Trường Đại học Ngoại thương", []string{"Đại học Ngoại thương", "FTU"}},
	{"TMU", "Trường Đại học Thương mại", []string{"Đại học Thương mại", "TMU"}},
	{"UTC", "Trường Đại học Giao thông Vận tải", []string{"Đại học Giao thông Vận tải", "UTC"}},
	{"NUCE", "Đại học Xây dựng Hà Nội", []string{"Trường Đại học Xây dựng Hà Nội", "Đại học Xây dựng"}},
	{"HAU", "Trường Đại học Kiến trúc Hà Nội", []string{"Đại học Kiến trúc Hà Nội", "HAU"}},
	{"HUMG", "Trường Đại học Mỏ - Địa chất", []string{"Đại học Mỏ Địa chất", "HUMG"}},
	{"EPU", "Trường Đại học Điện lực", []string{"Đại học Điện lực", "EPU"}},
	{"HAUI", "Trường Đại học Công nghiệp Hà Nội", []string{"Đại học Công nghiệp Hà Nội", "HAUI"}},
	{"UNETI", "Trường Đại học Kinh tế - Kỹ thuật Công nghiệp", []string{"Đại học Kinh tế Kỹ thuật Công nghiệp", "UNETI"}},
	{"UTC2", "Trường Đại học Công nghệ Giao thông Vận tải", []string{"Đại học Công nghệ GTVT", "UTT"}},
	{"HANU", "Trường Đại học Hà Nội", []string{"Đại học Hà Nội", "HANU"}},
	{"HLU", "Trường Đại học Luật Hà Nội", []string{"Đại học Luật Hà Nội", "HLU"}},
	{"HMU", "Trường Đại học Y Hà Nội", []string{"Đại học Y Hà Nội", "HMU"}},
	{"HUP", "Trường Đại học Dược Hà Nội", []string{"Đại học Dược Hà Nội", "HUP"}},
	{"HUPH", "Trường Đại học Y tế Công cộng", []string{"Đại học Y tế Công cộng", "HUPH"}},
	{"HNMU", "Trường Đại học Thủ đô Hà Nội", []string{"Đại học Thủ đô Hà Nội"}},
	{"HNUE", "Trường Đại học Sư phạm Hà Nội", []string{"Đại học Sư phạm Hà Nội", "HNUE"}},
	{"HPU2", "Trường Đại học Sư phạm Hà Nội 2", []string{"Đại học Sư phạm Hà Nội 2", "HPU2"}},
	{"NUAE", "Trường Đại học Sư phạm Nghệ thuật Trung ương", []string{"Đại học Sư phạm Nghệ thuật Trung ương", "NUAE"}},
	{"HUPES", "Trường Đại học Sư phạm Thể dục Thể thao Hà Nội", []string{"Đại học Sư phạm TDTT Hà Nội"}},
	{"UC", "Trường Đại học Công đoàn", []string{"Đại học Công đoàn"}},
	{"ULSA", "Trường Đại học Lao động - Xã hội", []string{"Đại học Lao động Xã hội", "ULSA"}},
	{"VNUA", "Học viện Nông nghiệp Việt Nam", []string{"Đại học Nông nghiệp Hà Nội", "VNUA"}},
	{"BA", "Học viện Ngân hàng", []string{"Banking Academy", "Học viện Ngân hàng Việt Nam"}},
	{"AOF", "Học viện Tài chính", []string{"Academy of Finance", "AOF"}},
	{"PTIT", "Học viện Công nghệ Bưu chính Viễn thông", []string{"Bưu chính Viễn thông", "PTIT"}},
	{"DAV", "Học viện Ngoại giao", []string{"Diplomatic Academy of Vietnam", "DAV"}},
	{"AJC", "Học viện Báo chí và Tuyên truyền", []string{"Báo chí Tuyên truyền", "AJC"}},
	{"APD", "Học viện Chính sách và Phát triển", []string{"Academy of Policy and Development", "APD"}},
	{"VWA", "Học viện Phụ nữ Việt Nam", []string{"Vietnam Women's Academy"}},
	{"VYA", "Học viện Thanh thiếu niên Việt Nam", nil},
	{"NAEM", "Học viện Quản lý Giáo dục", nil},
	{"ACTVN", "Học viện Kỹ thuật Mật mã", []string{"Kỹ thuật Mật mã", "ACTVN"}},
	{"VUTM", "Học viện Y Dược học Cổ truyền Việt Nam", nil},
	{"USTH", "Trường Đại học Khoa học và Công nghệ Hà Nội", []string{"Đại học Việt Pháp", "USTH"}},
	{"FPT", "Trường Đại học FPT", []string{"Đại học FPT", "FPT University"}},
	{"TLU", "Trường Đại học Thăng Long", []string{"Đại học Thăng Long"}},
	{"Phenikaa", "Trường Đại học Phenikaa", []string{"Đại học Phenikaa"}},
	{"DNU", "Trường Đại học Đại Nam", []string{"Đại học Đại Nam"}},
	{"HBU", "Trường Đại học Hòa Bình", []string{"Đại học Hòa Bình"}},
	{"HUBT", "Trường Đại học Kinh doanh và Công nghệ Hà Nội", []string{"Đại học Kinh doanh Công nghệ Hà Nội", "HUBT"}},
	{"", "Trường Đại học Công nghệ và Quản lý Hữu Nghị", []string{"Đại học Hữu Nghị"}},
	{"FBU", "Trường Đại học Tài chính - Ngân hàng Hà Nội", nil},
	{"PA", "Học viện An ninh Nhân dân", nil},
	{"PPA", "Học viện Cảnh sát Nhân dân", nil},
	{"FPU", "Trường Đại học Phòng cháy Chữa cháy", nil},
	{"MTA", "Học viện Kỹ thuật Quân sự", nil},
	{"VMMU", "Học viện Quân y", nil},
	{"VLA", "Học viện Khoa học Quân sự", nil},
	{"VAA", "Học viện Hàng không Việt Nam", []string{"Vietnam Aviation Academy"}},

	{"VNU-HCM", "Đại học Quốc gia Thành phố Hồ Chí Minh", []string{"ĐHQG TP.HCM", "ĐHQG HCM", "VNUHCM"}},
	{"HCMUT", "Trường Đại học Bách khoa - Đại học Quốc gia TP.HCM", []string{"Đại học Bách khoa ĐHQG TP.HCM", "Bách khoa TP.HCM", "HCMUT"}},
	{"UIT", "Trường Đại học Công nghệ Thông tin - Đại học Quốc gia TP.HCM", []string{"Đại học CNTT ĐHQG TP.HCM", "UIT"}},
	{"HCMUS", "Trường Đại học Khoa học Tự nhiên - Đại học Quốc gia TP.HCM", []string{"Đại học Khoa học Tự nhiên ĐHQG TP.HCM", "HCMUS"}},
	{"USSH", "Trường Đại học Khoa học Xã hội và Nhân văn - Đại học Quốc gia TP.HCM", []string{"Đại học KHXH&NV ĐHQG TP.HCM", "USSH TP.HCM"}},
	{"UEL", "Trường Đại học Kinh tế - Luật - Đại học Quốc gia TP.HCM", []string{"Đại học Kinh tế Luật ĐHQG TP.HCM", "UEL"}},
	{"HCMIU", "Trường Đại học Quốc tế - Đại học Quốc gia TP.HCM", []string{"Đại học Quốc tế ĐHQG TP.HCM", "IU", "HCMIU"}},
	{"UHS", "Trường Đại học Khoa học Sức khỏe - Đại học Quốc gia TP.HCM", []string{"Đại học Khoa học Sức khỏe ĐHQG TP.HCM", "UHS"}},
	{"UEH", "Đại học Kinh tế Thành phố Hồ Chí Minh", []string{"Trường Đại học Kinh tế TP.HCM", "Đại học Kinh tế TP.HCM", "UEH"}},
	{"HCMULAW", "Trường Đại học Luật TP.HCM", []string{"Đại học Luật TP.HCM", "ULAW"}},
	{"HCMUE", "Trường Đại học Sư phạm TP.HCM", []string{"Đại học Sư phạm TP.HCM", "HCMUE"}},
	{"HCMUTE", "Trường Đại học Công nghệ Kỹ thuật TP.HCM", []string{"Trường Đại học Sư phạm Kỹ thuật TP.HCM", "Đại học Sư phạm Kỹ thuật TP.HCM", "HCMUTE"}},
	{"IUH", "Trường Đại học Công nghiệp TP.HCM", []string{"Đại học Công nghiệp TP.HCM", "IUH"}},
	{"HUIT", "Trường Đại học Công Thương TP.HCM", []string{"Đại học Công nghiệp Thực phẩm TP.HCM", "HUFI", "HUIT"}},
	{"UFM", "Trường Đại học Tài chính - Marketing", []string{"Đại học Tài chính Marketing", "UFM"}},
	{"HUB", "Trường Đại học Ngân hàng TP.HCM", []string{"Đại học Ngân hàng TP.HCM", "HUB"}},
	{"HCMUAF", "Trường Đại học Nông Lâm TP.HCM", []string{"Đại học Nông Lâm TP.HCM", "NLU"}},
	{"UTH", "Trường Đại học Giao thông Vận tải TP.HCM", []string{"Đại học GTVT TP.HCM", "UTH"}},
	{"UAH", "Trường Đại học Kiến trúc TP.HCM", []string{"Đại học Kiến trúc TP.HCM", "UAH"}},
	{"HCMUNRE", "Trường Đại học Tài nguyên và Môi trường TP.HCM", []string{"Đại học Tài nguyên Môi trường TP.HCM"}},
	{"SGU", "Trường Đại học Sài Gòn", []string{"Đại học Sài Gòn", "SGU"}},
	{"UMP", "Trường Đại học Y Dược TP.HCM", []string{"Đại học Y Dược TP.HCM", "UMP"}},
	{"PNTU", "Trường Đại học Y khoa Phạm Ngọc Thạch", []string{"Đại học Y khoa Phạm Ngọc Thạch", "PNTU"}},
	{"VHS", "Trường Đại học Văn hóa TP.HCM", []string{"Đại học Văn hóa TP.HCM"}},
	{"HCMUFA", "Trường Đại học Mỹ thuật TP.HCM", []string{"Đại học Mỹ thuật TP.HCM"}},
	{"HCMCONS", "Nhạc viện Thành phố Hồ Chí Minh", []string{"Nhạc viện TP.HCM"}},
	{"HCMTDTT", "Trường Đại học Thể dục Thể thao TP.HCM", nil},
	{"HCMUPE", "Trường Đại học Sư phạm Thể dục Thể thao TP.HCM", nil},
	{"SKDA", "Trường Đại học Sân khấu - Điện ảnh TP.HCM", nil},
	{"TDTU", "Trường Đại học Tôn Đức Thắng", []string{"Đại học Tôn Đức Thắng", "TDTU"}},
	{"OU", "Trường Đại học Mở TP.HCM", []string{"Đại học Mở TP.HCM", "OU"}},
	{"HUTECH", "Trường Đại học Công nghệ TP.HCM", []string{"Đại học Công nghệ TP.HCM", "HUTECH"}},
	{"VLU", "Trường Đại học Văn Lang", []string{"Đại học Văn Lang", "VLU"}},
	{"HSU", "Trường Đại học Hoa Sen", []string{"Đại học Hoa Sen", "HSU"}},
	{"UEF", "Trường Đại học Kinh tế - Tài chính TP.HCM", []string{"Đại học UEF", "UEF"}},
	{"HIU", "Trường Đại học Quốc tế Hồng Bàng", []string{"Đại học Quốc tế Hồng Bàng", "HIU"}},
	{"NTTU", "Trường Đại học Nguyễn Tất Thành", []string{"Đại học Nguyễn Tất Thành", "NTTU"}},
	{"VHU", "Trường Đại học Văn Hiến", []string{"Đại học Văn Hiến", "VHU"}},
	{"SIU", "Trường Đại học Quốc tế Sài Gòn", []string{"Đại học Quốc tế Sài Gòn", "SIU"}},
	{"STU", "Trường Đại học Công nghệ Sài Gòn", []string{"Đại học Công nghệ Sài Gòn", "STU"}},
	{"HUFLIT", "Trường Đại học Ngoại ngữ - Tin học TP.HCM", []string{"Đại học HUFLIT", "HUFLIT"}},
	{"GDU", "Trường Đại học Gia Định", []string{"Đại học Gia Định", "GDU"}},
	{"HUNG-VUONG-HCM", "Trường Đại học Hùng Vương TP.HCM", []string{"Đại học Hùng Vương TP.HCM"}},
	{"RMIT", "Đại học RMIT Việt Nam", []string{"RMIT Việt Nam", "RMIT University Vietnam"}},
	{"FULBRIGHT", "Trường Đại học Fulbright Việt Nam", []string{"Fulbright University Vietnam", "FUV"}},

	{"UDN", "Đại học Đà Nẵng", []string{"University of Danang", "UDN"}},
	{"DUT", "Trường Đại học Bách khoa - Đại học Đà Nẵng", []string{"Đại học Bách khoa Đà Nẵng", "DUT"}},
	{"DUE", "Trường Đại học Kinh tế - Đại học Đà Nẵng", []string{"Đại học Kinh tế Đà Nẵng", "DUE"}},
	{"UED", "Trường Đại học Sư phạm - Đại học Đà Nẵng", []string{"Đại học Sư phạm Đà Nẵng", "UED"}},
	{"UFL", "Trường Đại học Ngoại ngữ - Đại học Đà Nẵng", []string{"Đại học Ngoại ngữ Đà Nẵng", "UFL"}},
	{"UTE", "Trường Đại học Sư phạm Kỹ thuật - Đại học Đà Nẵng", []string{"Đại học Sư phạm Kỹ thuật Đà Nẵng"}},
	{"VKU", "Trường Đại học Công nghệ Thông tin và Truyền thông Việt - Hàn - Đại học Đà Nẵng", []string{"VKU", "Đại học Việt Hàn"}},
	{"UD-UMP", "Trường Y Dược - Đại học Đà Nẵng", []string{"Y Dược Đại học Đà Nẵng"}},
	{"DUMTP", "Trường Đại học Kỹ thuật Y - Dược Đà Nẵng", []string{"Đại học Kỹ thuật Y Dược Đà Nẵng"}},
	{"DUFS", "Trường Đại học Thể dục Thể thao Đà Nẵng", nil},
	{"DTU", "Đại học Duy Tân", []string{"Trường Đại học Duy Tân", "DTU"}},
	{"DAU", "Trường Đại học Đông Á", []string{"Đại học Đông Á", "DAU"}},
	{"AUV", "Trường Đại học Mỹ tại Việt Nam", []string{"American University in Vietnam", "AUV"}},

	{"HU", "Đại học Huế", []string{"Hue University", "HU"}},
	{"HUSC", "Trường Đại học Khoa học - Đại học Huế", []string{"Đại học Khoa học Huế", "HUSC"}},
	{"HUCE", "Trường Đại học Sư phạm - Đại học Huế", []string{"Đại học Sư phạm Huế"}},
	{"HUE", "Trường Đại học Kinh tế - Đại học Huế", []string{"Đại học Kinh tế Huế"}},
	{"HUFL", "Trường Đại học Ngoại ngữ - Đại học Huế", []string{"Đại học Ngoại ngữ Huế"}},
	{"HUAF", "Trường Đại học Nông Lâm - Đại học Huế", []string{"Đại học Nông Lâm Huế"}},
	{"HUMP", "Trường Đại học Y - Dược - Đại học Huế", []string{"Đại học Y Dược Huế"}},
	{"HUL", "Trường Đại học Luật - Đại học Huế", []string{"Đại học Luật Huế"}},
	{"HUT", "Trường Du lịch - Đại học Huế", []string{"Đại học Du lịch Huế"}},
	{"HUFA", "Trường Đại học Nghệ thuật - Đại học Huế", []string{"Đại học Nghệ thuật Huế"}},

	{"TNU", "Đại học Thái Nguyên", []string{"Thai Nguyen University", "TNU"}},
	{"TNUT", "Trường Đại học Kỹ thuật Công nghiệp - Đại học Thái Nguyên", []string{"Đại học Kỹ thuật Công nghiệp Thái Nguyên", "TNUT"}},
	{"TUAF", "Trường Đại học Nông Lâm - Đại học Thái Nguyên", []string{"Đại học Nông Lâm Thái Nguyên", "TUAF"}},
	{"TNUE", "Trường Đại học Sư phạm - Đại học Thái Nguyên", []string{"Đại học Sư phạm Thái Nguyên", "TNUE"}},
	{"TUEBA", "Trường Đại học Kinh tế và Quản trị Kinh doanh - Đại học Thái Nguyên", []string{"Đại học Kinh tế QTKD Thái Nguyên", "TUEBA"}},
	{"TUMP", "Trường Đại học Y - Dược - Đại học Thái Nguyên", []string{"Đại học Y Dược Thái Nguyên"}},
	{"ICTU", "Trường Đại học Công nghệ Thông tin và Truyền thông - Đại học Thái Nguyên", []string{"Đại học CNTT&TT Thái Nguyên", "ICTU"}},
	{"TNU-S", "Trường Đại học Khoa học - Đại học Thái Nguyên", []string{"Đại học Khoa học Thái Nguyên"}},

	{"CTU", "Trường Đại học Cần Thơ", []string{"Đại học Cần Thơ", "CTU"}},
	{"CTUT", "Trường Đại học Kỹ thuật - Công nghệ Cần Thơ", []string{"Đại học Kỹ thuật Công nghệ Cần Thơ", "CTUT"}},
	{"CTUMP", "Trường Đại học Y Dược Cần Thơ", []string{"Đại học Y Dược Cần Thơ", "CTUMP"}},
	{"TDU", "Trường Đại học Tây Đô", []string{"Đại học Tây Đô"}},
	{"NCTU", "Trường Đại học Nam Cần Thơ", []string{"Đại học Nam Cần Thơ", "NCTU"}},
	{"FPT-CT", "Trường Đại học FPT - Cơ sở Cần Thơ", []string{"FPT Cần Thơ"}},

	{"QNU", "Trường Đại học Quy Nhơn", []string{"Đại học Quy Nhơn", "QNU"}},
	{"PYU", "Trường Đại học Phú Yên", []string{"Đại học Phú Yên"}},
	{"NTU", "Trường Đại học Nha Trang", []string{"Đại học Nha Trang", "NTU"}},
	{"TBD", "Trường Đại học Thái Bình Dương", []string{"Đại học Thái Bình Dương"}},
	{"PDU", "Trường Đại học Phạm Văn Đồng", []string{"Đại học Phạm Văn Đồng"}},
	{"UFA", "Trường Đại học Tài chính - Kế toán", []string{"Đại học Tài chính Kế toán"}},
	{"MUC", "Trường Đại học Xây dựng Miền Trung", []string{"Đại học Xây dựng Miền Trung"}},
	{"YU", "Trường Đại học Yersin Đà Lạt", []string{"Đại học Yersin Đà Lạt", "Yersin University"}},
	{"TTU", "Trường Đại học Tân Tạo", []string{"Đại học Tân Tạo", "TTU"}},

	{"", "Trường Đại học Tây Nguyên", []string{"Đại học Tây Nguyên", "TNU Tây Nguyên"}},
	{"VFU", "Trường Đại học Lâm nghiệp", []string{"Đại học Lâm nghiệp", "VFU"}},
	{"VNUF-GD", "Trường Đại học Lâm nghiệp - Phân hiệu Đồng Nai", []string{"Đại học Lâm nghiệp Đồng Nai"}},
	{"HDU", "Trường Đại học Hồng Đức", []string{"Đại học Hồng Đức", "HDU"}},
	{"TBU", "Trường Đại học Tây Bắc", []string{"Đại học Tây Bắc"}},
	{"HVU", "Trường Đại học Hùng Vương", []string{"Đại học Hùng Vương Phú Thọ", "HVU"}},
	{"TQU", "Trường Đại học Tân Trào", []string{"Đại học Tân Trào"}},
	{"SOU", "Trường Đại học Sao Đỏ", []string{"Đại học Sao Đỏ"}},
	{"HPU", "Trường Đại học Hải Phòng", []string{"Đại học Hải Phòng"}},
	{"VMU", "Trường Đại học Hàng hải Việt Nam", []string{"Đại học Hàng hải", "VMU"}},
	{"HMTU", "Trường Đại học Kỹ thuật Y tế Hải Dương", []string{"Đại học Kỹ thuật Y tế Hải Dương"}},
	{"HDIU", "Trường Đại học Hải Dương", []string{"Đại học Hải Dương"}},
	{"NDUN", "Trường Đại học Điều dưỡng Nam Định", []string{"Đại học Điều dưỡng Nam Định"}},
	{"NUTE", "Trường Đại học Sư phạm Kỹ thuật Nam Định", []string{"Đại học Sư phạm Kỹ thuật Nam Định"}},
	{"TBUMP", "Trường Đại học Y Dược Thái Bình", []string{"Đại học Y Thái Bình", "Đại học Y Dược Thái Bình"}},
	{"TBU-THAIBINH", "Trường Đại học Thái Bình", []string{"Đại học Thái Bình"}},
	{"HNU", "Trường Đại học Hoa Lư", []string{"Đại học Hoa Lư"}},
	{"HTU", "Trường Đại học Hà Tĩnh", []string{"Đại học Hà Tĩnh"}},
	{"QBU", "Trường Đại học Quảng Bình", []string{"Đại học Quảng Bình"}},
	{"VinhUni", "Trường Đại học Vinh", []string{"Đại học Vinh", "Vinh University"}},
	{"VUTED", "Trường Đại học Sư phạm Kỹ thuật Vinh", []string{"Đại học Sư phạm Kỹ thuật Vinh"}},
	{"VMU-VINH", "Trường Đại học Y khoa Vinh", []string{"Đại học Y khoa Vinh"}},
	{"NAUE", "Trường Đại học Kinh tế Nghệ An", []string{"Đại học Kinh tế Nghệ An"}},
	{"TDV", "Trường Đại học Văn hóa, Thể thao và Du lịch Thanh Hóa", []string{"Đại học Văn hóa Thể thao Du lịch Thanh Hóa"}},

	{"AGU", "Trường Đại học An Giang - Đại học Quốc gia TP.HCM", []string{"Đại học An Giang", "AGU"}},
	{"TVU", "Trường Đại học Trà Vinh", []string{"Đại học Trà Vinh", "TVU"}},
	{"DThU", "Trường Đại học Đồng Tháp", []string{"Đại học Đồng Tháp", "DThU"}},
	{"TGU", "Trường Đại học Tiền Giang", []string{"Đại học Tiền Giang", "TGU"}},
	{"KGU", "Trường Đại học Kiên Giang", []string{"Đại học Kiên Giang", "KGU"}},
	{"BLU", "Trường Đại học Bạc Liêu", []string{"Đại học Bạc Liêu"}},
	{"BDU", "Trường Đại học Bình Dương", []string{"Đại học Bình Dương"}},
	{"TDMU", "Trường Đại học Thủ Dầu Một", []string{"Đại học Thủ Dầu Một", "TDMU"}},
	{"EIU", "Trường Đại học Quốc tế Miền Đông", []string{"Đại học Quốc tế Miền Đông", "EIU"}},
	{"VGU", "Trường Đại học Việt Đức", []string{"Đại học Việt Đức", "Vietnamese-German University", "VGU"}},
	{"LHU", "Trường Đại học Lạc Hồng", []string{"Đại học Lạc Hồng", "LHU"}},
	{"DNU-DN", "Trường Đại học Đồng Nai", []string{"Đại học Đồng Nai"}},
	{"BRVT", "Trường Đại học Bà Rịa - Vũng Tàu", []string{"Đại học Bà Rịa Vũng Tàu", "BVU"}},
	{"PVU", "Trường Đại học Dầu khí Việt Nam", []string{"Đại học Dầu khí Việt Nam", "PVU"}},

	{"BMTU", "Trường Đại học Buôn Ma Thuột", []string{"Đại học Buôn Ma Thuột", "BMTU"}},
	{"YDU", "Trường Đại học Y Dược Buôn Ma Thuột", []string{"Đại học Y Dược Buôn Ma Thuột"}},
	{"QNUI", "Trường Đại học Quảng Nam", []string{"Đại học Quảng Nam"}},
	{"KONTUM", "Phân hiệu Đại học Đà Nẵng tại Kon Tum", []string{"Đại học Đà Nẵng Kon Tum", "UD-CK"}},

	{"ANU", "Trường Đại học An ninh Nhân dân", nil},
	{"PPU", "Trường Đại học Cảnh sát Nhân dân", nil},
	{"TDTN", "Trường Đại học Trần Đại Nghĩa", nil},
	{"NQU", "Trường Đại học Ngô Quyền", []string{"Trường Sĩ quan Công binh"}},
	{"LQO", "Trường Đại học Trần Quốc Tuấn", []string{"Trường Sĩ quan Lục quân 1"}},
}

var CommonMajors = []string{
	"Công nghệ thông tin", "Khoa học máy tính", "Kỹ thuật phần mềm", "Hệ thống thông tin", "Mạng máy tính và truyền thông dữ liệu", "An toàn thông tin", "Trí tuệ nhân tạo", "Khoa học dữ liệu", "Kỹ thuật máy tính", "Công nghệ kỹ thuật máy tính",
	"Điện - Điện tử", "Kỹ thuật điện", "Kỹ thuật điện tử - viễn thông", "Kỹ thuật điều khiển và tự động hóa", "Kỹ thuật cơ điện tử", "Kỹ thuật cơ khí", "Kỹ thuật ô tô", "Kỹ thuật nhiệt", "Kỹ thuật xây dựng", "Kiến trúc", "Quy hoạch vùng và đô thị",
	"Công nghệ sinh học", "Công nghệ thực phẩm", "Kỹ thuật hóa học", "Hóa học", "Vật lý học", "Toán học", "Toán ứng dụng", "Khoa học môi trường", "Quản lý tài nguyên và môi trường", "Nông học", "Chăn nuôi", "Thú y", "Lâm nghiệp", "Nuôi trồng thủy sản",
	"Y khoa", "Răng - Hàm - Mặt", "Dược học", "Điều dưỡng", "Y học cổ truyền", "Y học dự phòng", "Kỹ thuật xét nghiệm y học", "Kỹ thuật hình ảnh y học", "Dinh dưỡng", "Y tế công cộng",
	"Quản trị kinh doanh", "Kinh doanh quốc tế", "Marketing", "Thương mại điện tử", "Kinh tế", "Kinh tế quốc tế", "Tài chính - Ngân hàng", "Kế toán", "Kiểm toán", "Logistics và quản lý chuỗi cung ứng", "Quản trị nhân lực", "Quản trị văn phòng", "Bất động sản", "Bảo hiểm",
	"Luật", "Luật kinh tế", "Luật quốc tế", "Quản lý nhà nước", "Quan hệ quốc tế", "Chính trị học", "Xã hội học", "Tâm lý học", "Công tác xã hội", "Đông phương học", "Việt Nam học", "Quốc tế học",
	"Báo chí", "Truyền thông đa phương tiện", "Quan hệ công chúng", "Ngôn ngữ Anh", "Ngôn ngữ Trung Quốc", "Ngôn ngữ Nhật", "Ngôn ngữ Hàn Quốc", "Ngôn ngữ Pháp", "Ngôn ngữ Nga", "Văn học", "Lịch sử", "Địa lý học",
	"Du lịch", "Quản trị dịch vụ du lịch và lữ hành", "Quản trị khách sạn", "Quản trị nhà hàng và dịch vụ ăn uống", "Thiết kế đồ họa", "Thiết kế thời trang", "Thiết kế nội thất", "Mỹ thuật", "Âm nhạc", "Điện ảnh - Truyền hình",
	"Giáo dục Mầm non", "Giáo dục Tiểu học", "Sư phạm Toán học", "Sư phạm Tin học", "Sư phạm Vật lý", "Sư phạm Hóa học", "Sư phạm Sinh học", "Sư phạm Ngữ văn", "Sư phạm Lịch sử", "Sư phạm Địa lý", "Sư phạm Tiếng Anh", "Giáo dục thể chất",
}

func init() {
	sort.SliceStable(VietnamSchools, func(i, j int) bool { return VietnamSchools[i].Name < VietnamSchools[j].Name })
	sort.Strings(CommonMajors)
}

func normalizeText(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	repl := strings.NewReplacer(
		"à", "a", "á", "a", "ạ", "a", "ả", "a", "ã", "a", "â", "a", "ầ", "a", "ấ", "a", "ậ", "a", "ẩ", "a", "ẫ", "a", "ă", "a", "ằ", "a", "ắ", "a", "ặ", "a", "ẳ", "a", "ẵ", "a",
		"è", "e", "é", "e", "ẹ", "e", "ẻ", "e", "ẽ", "e", "ê", "e", "ề", "e", "ế", "e", "ệ", "e", "ể", "e", "ễ", "e",
		"ì", "i", "í", "i", "ị", "i", "ỉ", "i", "ĩ", "i",
		"ò", "o", "ó", "o", "ọ", "o", "ỏ", "o", "õ", "o", "ô", "o", "ồ", "o", "ố", "o", "ộ", "o", "ổ", "o", "ỗ", "o", "ơ", "o", "ờ", "o", "ớ", "o", "ợ", "o", "ở", "o", "ỡ", "o",
		"ù", "u", "ú", "u", "ụ", "u", "ủ", "u", "ũ", "u", "ư", "u", "ừ", "u", "ứ", "u", "ự", "u", "ử", "u", "ữ", "u",
		"ỳ", "y", "ý", "y", "ỵ", "y", "ỷ", "y", "ỹ", "y", "đ", "d",
	)
	value = repl.Replace(value)
	value = strings.NewReplacer("đh ", "dai hoc ", "dh ", "dai hoc ", "hv ", "hoc vien ", "cd ", "cao dang ", "tp.hcm", "tp hcm", "tp. hcm", "tp hcm", "tp hồ chí minh", "tp hcm", "thành phố hồ chí minh", "tp hcm").Replace(value)
	var b strings.Builder
	lastSpace := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
		} else if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func compactSchoolKey(value string) string {
	value = normalizeText(value)
	prefixes := []string{"truong dai hoc ", "dai hoc ", "hoc vien ", "truong cao dang ", "cao dang ", "truong "}
	for _, p := range prefixes {
		if strings.HasPrefix(value, p) {
			value = strings.TrimSpace(strings.TrimPrefix(value, p))
			break
		}
	}
	return value
}

func schoolKeys(s School) []string {
	vals := []string{s.Name, s.Code}
	vals = append(vals, s.Aliases...)
	out := make([]string, 0, len(vals)*2)
	seen := map[string]bool{}
	for _, raw := range vals {
		for _, key := range []string{normalizeText(raw), compactSchoolKey(raw)} {
			if key != "" && !seen[key] {
				seen[key] = true
				out = append(out, key)
			}
		}
	}
	return out
}

func CanonicalSchool(value string) (string, bool) {
	key := normalizeText(value)
	core := compactSchoolKey(value)
	if key == "" {
		return "", true
	}
	var candidate string
	for _, school := range VietnamSchools {
		matched := false
		for _, k := range schoolKeys(school) {
			if key == k || core == k {
				matched = true
				break
			}
		}
		if matched {
			if candidate != "" && candidate != school.Name {
				return "", false
			}
			candidate = school.Name
		}
	}
	if candidate == "" {
		return "", false
	}
	return candidate, true
}

func CanonicalMajor(value string) string {
	key := normalizeText(value)
	if key == "" {
		return ""
	}
	for _, major := range CommonMajors {
		if normalizeText(major) == key {
			return major
		}
	}
	return strings.TrimSpace(value)
}

func SameSchool(a, b string) bool {
	ca, oka := CanonicalSchool(a)
	cb, okb := CanonicalSchool(b)
	if oka && okb && ca != "" && cb != "" {
		return ca == cb
	}
	// Backward compatibility for old records not yet present in the catalogue.
	return normalizeText(a) != "" && normalizeText(a) == normalizeText(b)
}

package reactions

import "time"

type Summary struct {
	PostSlug      string `json:"postSlug"`
	LikeCount     int    `json:"likeCount"`
	DislikeCount  int    `json:"dislikeCount"`
	BookmarkCount int    `json:"bookmarkCount"`
	MyReaction    string `json:"myReaction"`
	Bookmarked    bool   `json:"bookmarked"`
}

type ReactionRequest struct {
	Type string `json:"type"`
}

type BookmarkRequest struct {
	Bookmarked bool `json:"bookmarked"`
}

type Bookmark struct {
	PostSlug     string    `json:"postSlug"`
	BookmarkedAt time.Time `json:"bookmarkedAt"`
}

type BookmarkQuery struct {
	Keyword  string
	Category string
	Sort     string
	Page     int
	PageSize int
}

type BookmarkPage struct {
	Items    []Bookmark
	Page     int
	PageSize int
	Total    int
}

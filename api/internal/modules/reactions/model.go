package reactions

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

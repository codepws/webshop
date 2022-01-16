package redis

/*
	Redis Key
*/
const (
	KeyPrefix                 = "bluebell:"
	KeyPostTimeZSet           = "bluebell:post:time"
	KeyPostScoreZSet          = "bluebell:post:score"
	KeyPostVoteZSetPrefix     = "bluebell:post:vote:"
	KeyCommunityPostSetPrefix = "bluebell:community:"
)

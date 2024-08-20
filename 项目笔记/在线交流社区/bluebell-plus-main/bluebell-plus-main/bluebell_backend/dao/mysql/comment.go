package mysql

import (
	"bluebell_backend/models"

	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

func CreateComment(comment *models.Comment) (err error) {
	sqlStr := `insert into comment(
	comment_id, content, post_id, author_id, parent_id)
	values(?,?,?,?,?)`
	_, err = db.Exec(sqlStr, comment.CommentID, comment.Content, comment.PostID,
		comment.AuthorID, comment.ParentID)
	if err != nil {
		zap.L().Error("insert comment failed", zap.Error(err))
		err = ErrorInsertFailed
		return
	}
	return
}

func GetCommentListByIDs(post_id string) (commentList []*models.Comment, err error) {
	sqlStr := `select comment_id, content, post_id, author_id, parent_id, update_time
	from comment
	where post_id = ?`
	// 动态填充id
	// query, args, err := sqlx.In(sqlStr, post_id)
	// if err != nil {
	// 	return
	// }
	// // sqlx.In 返回带 `?` bindVar 的查询语句, 我们使用Rebind()重新绑定它
	// query = db.Rebind(query)
	err = db.Select(&commentList, sqlStr,post_id)
	return
}
